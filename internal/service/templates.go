package service

type Template struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"` // Emoji ou URL
	Description string `json:"description"`
	MinCPU      int    `json:"min_cpu"`
	MinRAM      int    `json:"min_ram_mb"` // MB
	CloudConfig string `json:"-"` // O YAML do cloud-init (não enviar no JSON de lista)
}

func GetTemplates() []Template {
	return []Template{
		{
			ID:          "docker-host",
			Name:        "Docker Host",
			Icon:        "🐳",
			Description: "Docker installation with curl and basic setup",
			MinCPU:      1,
			MinRAM:      1024, // 1GB
			CloudConfig: `#cloud-config
packages:
  - curl
runcmd:
  - curl -fsSL https://get.docker.com | sh
  - systemctl enable docker
  - systemctl start docker
users:
  - name: axion
    groups: [docker]
    shell: /bin/bash
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
    ssh_authorized_keys:
      - $AXION_SSH_KEY
`,
		},
		{
			ID:          "k3s-node",
			Name:        "K3s Node",
			Icon:        "🚢",
			Description: "Lightweight Kubernetes distribution",
			MinCPU:      2,
			MinRAM:      2048, // 2GB
			CloudConfig: `#cloud-config
packages:
  - curl
runcmd:
  - curl -sfL https://get.k3s.io | sh -
  - systemctl enable k3s
  - systemctl start k3s
users:
  - name: axion
    groups: [k3s]
    shell: /bin/bash
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
    ssh_authorized_keys:
      - $AXION_SSH_KEY
`,
		},
		{
			ID:          "minecraft-server",
			Name:        "Minecraft Server",
			Icon:        "⛏️",
			Description: "Minecraft server with Java installation",
			MinCPU:      2,
			MinRAM:      4096, // 4GB
			CloudConfig: `#cloud-config
packages:
  - openjdk-21-jre-headless
  - curl
runcmd:
  - mkdir -p /opt/minecraft
  - curl -o /opt/minecraft/server.jar https://piston-data.mojang.com/v1/objects/8f3112a104976e52abcd3d7516729e5e9f09d32e/server.jar
  - echo 'eula=true' > /opt/minecraft/eula.txt
  - chmod +x /opt/minecraft/server.jar
  - |
    cat << 'EOF' > /etc/systemd/system/minecraft.service
[Unit]
Description=Minecraft Server
After=network.target

[Service]
Type=simple
User=axion
WorkingDirectory=/opt/minecraft
ExecStart=/usr/bin/java -Xmx3G -Xms1G -jar server.jar nogui
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF
  - systemctl daemon-reload
  - systemctl enable minecraft
  - systemctl start minecraft
users:
  - name: axion
    shell: /bin/bash
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
    ssh_authorized_keys:
      - $AXION_SSH_KEY
`,
		},
		{
			ID:          "postgresql",
			Name:        "PostgreSQL Database",
			Icon:        "🗄️",
			Description: "PostgreSQL database server",
			MinCPU:      1,
			MinRAM:      1024, // 1GB
			CloudConfig: `#cloud-config
packages:
  - postgresql-16
  - postgresql-client-16
runcmd:
  - systemctl enable postgresql
  - systemctl start postgresql
  - sudo -u postgres psql -c "CREATE USER axion WITH PASSWORD 'axion';"
  - sudo -u postgres psql -c "ALTER USER axion CREATEDB;"
  - sudo -u postgres psql -c "CREATE DATABASE axion OWNER axion;"
  - sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/" /etc/postgresql/16/main/postgresql.conf
  - echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/16/main/pg_hba.conf
  - systemctl restart postgresql
users:
  - name: axion
    shell: /bin/bash
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
    ssh_authorized_keys:
      - $AXION_SSH_KEY
`,
		},
		{
			ID:          "nginx-web",
			Name:        "Nginx Web Server",
			Icon:        "🌐",
			Description: "Nginx web server with custom index page",
			MinCPU:      1,
			MinRAM:      512, // 512MB
			CloudConfig: `#cloud-config
packages:
  - nginx
  - curl
write_files:
  - path: /var/www/html/index.html
    content: |
      <!DOCTYPE html>
      <html>
      <head>
          <title>Welcome to Your Nginx Server</title>
          <style>
              body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
              h1 { color: #4a90e2; }
          </style>
      </head>
      <body>
          <h1>Welcome to Your Nginx Server!</h1>
          <p>Your Axion instance is running successfully.</p>
      </body>
      </html>
runcmd:
  - systemctl enable nginx
  - systemctl start nginx
users:
  - name: axion
    shell: /bin/bash
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
    ssh_authorized_keys:
      - $AXION_SSH_KEY
`,
		},
	}
}