package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"aexon/internal/api"
	"aexon/internal/auth"
	"aexon/internal/db"
	"aexon/internal/provider/lxc"
	"aexon/internal/types"
	"aexon/internal/utils"
	"aexon/internal/worker"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InstanceActionRequest struct {
	Action string `json:"action" binding:"required"`
}

type InstanceLimitsRequest struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}

type CreateInstanceRequest struct {
	Name     string            `json:"name" binding:"required"`
	Image    string            `json:"image" binding:"required"`
	Limits   map[string]string `json:"limits"`
	UserData string            `json:"user_data"` // Opcional: Cloud-Init
}

type SnapshotRequest struct {
	Name string `json:"name" binding:"required"`
}

type AddPortRequest struct {
	HostPort      int    `json:"host_port" binding:"required"`
	ContainerPort int    `json:"container_port" binding:"required"`
	Protocol      string `json:"protocol" binding:"required"`
}

func main() {
	log.SetOutput(os.Stdout)
	log.Println("Iniciando Axion Control Plane...")

	if err := db.Init("axion.db"); err != nil {
		log.Fatalf("[ERRO CRÍTICO] Falha ao inicializar banco de dados: %v", err)
	}
	log.Println("Database axion.db inicializado.")

	lxcClient, err := lxc.NewClient()
	if err != nil {
		log.Fatalf("[ERRO CRÍTICO] Falha na inicialização do provider LXD: %v", err)
	}
	log.Println("Conexão com LXD estabelecida.")

	worker.Init(2, lxcClient)
	api.InitBroadcaster()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "OPTIONS", "DELETE", "PUT"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:   []string{"Content-Length", "Content-Disposition"},
		MaxAge:          12 * time.Hour,
	}))

	r.POST("/login", auth.LoginHandler)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	protected := r.Group("/")
	protected.Use(auth.AuthMiddleware())
	{
		// Instances
		protected.GET("/instances", func(c *gin.Context) {
			instances, err := lxcClient.ListInstances()
			if err != nil {
				log.Printf("Erro ao processar ListInstances: %v", err)
				c.JSON(500, gin.H{"error": "Falha ao obter métricas"})
				return
			}
			c.JSON(200, instances)
		})

		protected.POST("/instances", func(c *gin.Context) {
			var req CreateInstanceRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "JSON inválido. Campos obrigatórios: name, image"})
				return
			}
			reqCpu := 1
			if val, ok := req.Limits["limits.cpu"]; ok { reqCpu = utils.ParseCpuCores(val) }
			reqRam := int64(512)
			if val, ok := req.Limits["limits.memory"]; ok { reqRam = utils.ParseMemoryToMB(val) }

			if err := lxcClient.CheckGlobalQuota(reqCpu, reqRam); err != nil {
				c.JSON(409, gin.H{"error": "Quota Exceeded", "details": err.Error()})
				return
			}

			jobID := uuid.New().String()
			payloadBytes, _ := json.Marshal(req)
			job := &db.Job{ID: jobID, Type: types.JobTypeCreateInstance, Target: req.Name, Payload: string(payloadBytes)}
			if err := db.CreateJob(job); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			worker.DispatchJob(jobID)
			c.JSON(202, gin.H{"job_id": jobID, "status": "accepted"})
		})

		protected.DELETE("/instances/:name", func(c *gin.Context) {
			name := c.Param("name")
			jobID := uuid.New().String()
			job := &db.Job{ID: jobID, Type: types.JobTypeDeleteInstance, Target: name, Payload: "{}"}
			if err := db.CreateJob(job); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			worker.DispatchJob(jobID)
			c.JSON(202, gin.H{"job_id": jobID, "status": "accepted"})
		})

		protected.POST("/instances/:name/state", func(c *gin.Context) {
			name := c.Param("name")
			var req InstanceActionRequest
			if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": "JSON inválido"}); return }
			jobID := uuid.New().String()
			payloadBytes, _ := json.Marshal(req)
			job := &db.Job{ID: jobID, Type: types.JobTypeStateChange, Target: name, Payload: string(payloadBytes)}
			if err := db.CreateJob(job); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			worker.DispatchJob(jobID)
			c.JSON(202, gin.H{"job_id": jobID, "status": "accepted"})
		})

		protected.POST("/instances/:name/limits", func(c *gin.Context) {
			name := c.Param("name")
			var req InstanceLimitsRequest
			if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": "JSON inválido"}); return }
			reqCpu := utils.ParseCpuCores(req.CPU)
			reqRam := utils.ParseMemoryToMB(req.Memory)
			if err := lxcClient.CheckGlobalQuota(reqCpu, reqRam); err != nil { c.JSON(409, gin.H{"error": "Quota Exceeded", "details": err.Error()}); return }
			jobID := uuid.New().String()
			payloadBytes, _ := json.Marshal(req)
			job := &db.Job{ID: jobID, Type: types.JobTypeUpdateLimits, Target: name, Payload: string(payloadBytes)}
			if err := db.CreateJob(job); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			worker.DispatchJob(jobID)
			c.JSON(202, gin.H{"job_id": jobID, "status": "accepted"})
		})

		// Snapshots
		protected.GET("/instances/:name/snapshots", func(c *gin.Context) {
			name := c.Param("name")
			snaps, err := lxcClient.ListSnapshots(name)
			if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, snaps)
		})
		protected.POST("/instances/:name/snapshots", func(c *gin.Context) {
			name := c.Param("name")
			var req SnapshotRequest
			if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": "Nome obrigatório"}); return }
			jobID := uuid.New().String()
			payload, _ := json.Marshal(map[string]string{"snapshot_name": req.Name})
			job := &db.Job{ID: jobID, Type: types.JobTypeCreateSnapshot, Target: name, Payload: string(payload)}
			if err := db.CreateJob(job); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			worker.DispatchJob(jobID)
			c.JSON(202, gin.H{"job_id": jobID, "status": "accepted"})
		})
		protected.POST("/instances/:name/snapshots/:snap/restore", func(c *gin.Context) {
			name := c.Param("name")
			snap := c.Param("snap")
			jobID := uuid.New().String()
			payload, _ := json.Marshal(map[string]string{"snapshot_name": snap})
			job := &db.Job{ID: jobID, Type: types.JobTypeRestoreSnapshot, Target: name, Payload: string(payload)}
			if err := db.CreateJob(job); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			worker.DispatchJob(jobID)
			c.JSON(202, gin.H{"job_id": jobID, "status": "accepted"})
		})
		protected.DELETE("/instances/:name/snapshots/:snap", func(c *gin.Context) {
			name := c.Param("name")
			snap := c.Param("snap")
			jobID := uuid.New().String()
			payload, _ := json.Marshal(map[string]string{"snapshot_name": snap})
			job := &db.Job{ID: jobID, Type: types.JobTypeDeleteSnapshot, Target: name, Payload: string(payload)}
			if err := db.CreateJob(job); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			worker.DispatchJob(jobID)
			c.JSON(202, gin.H{"job_id": jobID, "status": "accepted"})
		})

		// Ports
		protected.POST("/instances/:name/ports", func(c *gin.Context) {
			name := c.Param("name")
			var req AddPortRequest
			if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": err.Error()}); return }
			jobID := uuid.New().String()
			payload, _ := json.Marshal(req)
			job := &db.Job{ID: jobID, Type: types.JobTypeAddPort, Target: name, Payload: string(payload)}
			if err := db.CreateJob(job); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			worker.DispatchJob(jobID)
			c.JSON(202, gin.H{"job_id": jobID, "status": "accepted"})
		})
		protected.DELETE("/instances/:name/ports/:host_port", func(c *gin.Context) {
			name := c.Param("name")
			hpStr := c.Param("host_port")
			hp, _ := strconv.Atoi(hpStr)
			jobID := uuid.New().String()
			payload, _ := json.Marshal(map[string]int{"host_port": hp})
			job := &db.Job{ID: jobID, Type: types.JobTypeRemovePort, Target: name, Payload: string(payload)}
			if err := db.CreateJob(job); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			worker.DispatchJob(jobID)
			c.JSON(202, gin.H{"job_id": jobID, "status": "accepted"})
		})

		// --- FILE SYSTEM (EXPLORER) ---
		
		// List
		protected.GET("/instances/:name/files/list", func(c *gin.Context) {
			name := c.Param("name")
			path := c.Query("path")
			if path == "" { path = "/root" }

			entries, err := lxcClient.ListFiles(name, path)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, entries)
		})

		// Download
		protected.GET("/instances/:name/files/content", func(c *gin.Context) {
			name := c.Param("name")
			rawPath := c.Query("path")
			if rawPath == "" { c.JSON(400, gin.H{"error": "Path missing"}); return }

			cleanPath := filepath.Clean(rawPath) 

			content, size, err := lxcClient.DownloadFile(name, cleanPath)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			defer content.Close()

			const MaxEditorSize = 1024 * 1024 // 1MB

			if size > MaxEditorSize {
				// Force download ONLY for files confirmed to be large
				c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(cleanPath)))
			} else {
				c.Header("Content-Disposition", "inline")
			}

			if size > 0 {
				c.Header("Content-Length", fmt.Sprintf("%d", size))
			}
			io.Copy(c.Writer, content)
		})

		// Upload
		protected.POST("/instances/:name/files", func(c *gin.Context) {
			name := c.Param("name")
			path := c.Query("path")
			if path == "" { c.JSON(400, gin.H{"error": "Target path required"}); return }

			fileHeader, err := c.FormFile("file")
			if err != nil {
				c.JSON(400, gin.H{"error": "File missing"}); return
			}

			log.Printf("[Upload Debug] Filename: %s, Header Size: %d", fileHeader.Filename, fileHeader.Size)

			file, err := fileHeader.Open()
			if err != nil { c.JSON(500, gin.H{"error": "Failed to open file"}); return }
			defer file.Close()

			if err := lxcClient.UploadFile(name, path, file); err != nil {
				c.JSON(500, gin.H{"error": "Upload failed", "details": err.Error()})
				return
			}

			c.JSON(200, gin.H{"status": "uploaded"})
		})

		// Delete File
		protected.DELETE("/instances/:name/files", func(c *gin.Context) {
			name := c.Param("name")
			path := c.Query("path")
			if path == "" { c.JSON(400, gin.H{"error": "Path missing"}); return }

			if err := lxcClient.DeleteFile(name, path); err != nil {
				c.JSON(500, gin.H{"error": "Delete failed", "details": err.Error()})
				return
			}
			c.JSON(200, gin.H{"status": "deleted"})
		})

		protected.GET("/jobs", func(c *gin.Context) {
			jobs, err := db.ListRecentJobs(50)
			if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
			c.JSON(200, jobs)
		})
		protected.GET("/jobs/:id", func(c *gin.Context) {
			id := c.Param("id")
			job, err := db.GetJob(id)
			if err != nil { c.JSON(404, gin.H{"error": "Not found"}); return }
			c.JSON(200, job)
		})
		protected.GET("/ws/telemetry", func(c *gin.Context) {
			api.StreamTelemetry(c, lxcClient)
		})
	}

	r.GET("/ws/terminal/:name", func(c *gin.Context) {
		api.TerminalHandler(c, lxcClient)
	})

	port := "8500"
	log.Printf("Axion Control Plane rodando na porta %s", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Falha ao iniciar servidor web: %v", err)
	}
}