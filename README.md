# ⚡ Axion Control Plane

### HPC-First LXC Orchestration Platform

> **Axion** é uma plataforma de orquestração e virtualização focada em **performance extrema, alta densidade de containers e controle total do host**.
> Ele nasce com uma filosofia clara: **menos abstração, mais performance real**.

Diferente de soluções genéricas, o Axion é projetado para:

* Máxima eficiência por core
* Overhead praticamente zero
* Telemetria em tempo real
* Governança rígida de recursos
* Arquitetura assíncrona enterprise-grade

---

## 🚀 Visão do Projeto

O Axion foi criado com um objetivo direto:

> **Extrair o máximo absoluto de performance do hardware disponível usando LXC.**

Ele é ideal para:

* Game servers de alta densidade
* Ambientes de staging e produção
* Infraestrutura para SaaS
* Plataformas de CI/CD
* Laboratórios de desenvolvimento
* Ambientes educacionais
* Clusters de containers de alta performance

Nada de hipervisores pesados.
Nada de overengineering desnecessário.
Aqui, **cada ciclo de CPU importa**.

---

## 🧠 Filosofia do Axion

* **Performance acima de tudo**
* **Latência mínima**
* **Arquitetura enxuta**
* **Controle total do host**
* **Alta densidade por nó**
* **Automação nativa**
* **Sem vendor lock-in**

---

## ✅ Escopo Atual (v1.x)

Atualmente, o Axion é um **Control Plane completo para Containers LXC**, utilizando o LXD como runtime base.

### 🔹 O que o Axion é HOJE:

* Orquestrador LXC
* Painel Web em tempo real
* Job System assíncrono
* Governança global de recursos
* Autenticação por JWT
* Auditoria de ações
* Controle completo do ciclo de vida dos containers

### 🔹 O que o Axion NÃO é ainda:

* ❌ Orquestrador multi-node
* ❌ Hypervisor de VMs (KVM)
* ❌ Plataforma bare-metal
* ❌ Orquestrador de GPU

Esses pontos fazem parte da **v2.0+**.

---

## 🏗️ Arquitetura Atual (Implementada)

### 🔧 Backend (Control Plane)

* **Linguagem:** Go 1.22+
* **Framework HTTP:** Gin
* **Persistência:** SQLite (WAL Mode)
* **Autenticação:** JWT (24h)
* **Arquitetura:** Totalmente assíncrona via Jobs
* **Worker Pool:** 2 workers concorrentes
* **WebSocket:** Telemetria + Eventos de Jobs
* **Governança:** Quota global de CPU e RAM
* **Resiliência:**

  * Locks por container
  * Retry com backoff exponencial
  * Timeout por tipo de job
  * Recovery de jobs presos

---

### 📦 Runtime de Containers

* **Tecnologia:** LXC/LXD
* **Conexão:** Socket Unix direto
* **Operações:**

  * Create
  * Start
  * Stop
  * Restart
  * Update CPU/RAM
* **Telemetria:** CPU e RAM em tempo real (1s)

⚠️ Todos os containers **compartilham o kernel do host**, garantindo:

* Overhead mínimo
* Boot instantâneo
* Performance próxima ao bare-metal

---

### 🌐 Comunicação em Tempo Real

* WebSocket multiplexado:

  * Telemetria de containers
  * Eventos de Jobs (PENDING → IN_PROGRESS → COMPLETED/FAILED)
* Event Bus interno com fan-out

---

## 🖥️ Frontend (Painel Web)

* **Framework:** Next.js 14+ (App Router)
* **Design:** Enterprise Dark (Zinc + Indigo)
* **Features:**

  * Login com JWT
  * Dashboard com cards em tempo real
  * Gráficos sparkline de CPU/RAM
  * Controle Start/Stop/Restart
  * Wizard de criação de instâncias
  * Settings Panel para CPU/RAM
  * Activity Drawer com auditoria de Jobs
* **Segurança:**

  * Proteção de rotas
  * Redirecionamento automático para /login
  * Logout forçado ao receber 401

---

## 🛡️ Segurança

* Autenticação JWT
* Middleware para rotas e WebSocket
* Auditoria de Jobs
* Locks de execução por container
* Quota global de recursos
* Prevenção contra overcommit

---

## 📊 Governança de Recursos

* **Limite Global Atual:**

  * 8 vCPU
  * 8 GB RAM
* Validação antes de:

  * Criar containers
  * Atualizar limites
* Retorno semântico:

  * `409 Conflict` ao exceder capacidade

Isso impede que usuários:

* Travam o host
* Criem instâncias infinitas
* Inflacionem recursos sem controle

---

## 📦 Planos do Axion

### 🧪 Axion Personal

* Projetos pessoais
* Estudos
* Ambientes locais
* Sem SLA
* Comunidade

---

### 🏢 Axion Enterprise

* Uso comercial
* Suporte 24/7
* SLA garantido
* Auditoria avançada
* Backup corporativo
* Multi-ambiente
* Compliance (LGPD, ISO, etc.)

---

## 🧬 Roadmap

### ✅ v1.x (Atual)

* [x] Control Plane LXC
* [x] Painel Web
* [x] Telemetria em tempo real
* [x] Job System assíncrono
* [x] Governança de recursos
* [x] Autenticação JWT
* [x] Wizard de criação de instâncias

---

### 🚀 v2.0 (Futuro)

* [ ] Multi-node Control Plane
* [ ] dqlite ou etcd
* [ ] Suporte a VMs via LXD (KVM)
* [ ] Scheduler distribuído
* [ ] Quotas por usuário
* [ ] Orquestração de clusters

---

## 🛠️ Tecnologias

### ✅ Atuais

* Go
* Gin
* LXD / LXC
* SQLite (WAL)
* Next.js
* WebSocket
* JWT

### 🔮 Futuras

* KVM
* dqlite / etcd
* ZFS
* Ceph
* Kubernetes
* Slurm

---

## 📜 Licenciamento

O **Axion não é open-source completo**.

Modelo de licença:

* Uso pessoal
* Uso educacional
* Uso comercial
* Por cluster ou infraestrutura

Alguns módulos poderão ser abertos futuramente.

---

## ⚠️ Status Atual

> ✅ **Projeto ativo e funcional em produção local.**
> O Axion **já possui backend, frontend, job system, segurança, telemetria e governança de recursos implementados.**

---

## 🧠 Frase Oficial

> **“Axion não gerencia máquinas. Ele extrai o máximo do hardware.”**

