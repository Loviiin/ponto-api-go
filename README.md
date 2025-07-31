# API de Ponto Eletrônico (Versão em Go)

![Versão da Linguagem](https://img.shields.io/badge/go-1.21+-blue.svg) ![Framework](https://img.shields.io/badge/Gin-v1.9-cyan.svg) ![Licença](https://img.shields.io/badge/license-MIT-green.svg)

## 📖 Sobre o Projeto

A **API de Ponto Eletrônico (Go)** é um sistema backend de alta performance, projetado para gerenciar o registro de jornada de trabalho de funcionários. Este projeto foi desenvolvido como uma peça central de portfólio, demonstrando a aplicação de uma arquitetura limpa, boas práticas e segurança em um ambiente **Go (Golang)**.

Este projeto é a contraparte do [ponto-api (versão em Java/Spring)](<URL_PARA_SEU_REPO_JAVA_AQUI>), demonstrando a capacidade de resolver o mesmo problema com diferentes stacks de tecnologia, com foco em performance e eficiência de recursos, características marcantes do Go.

## ✨ Features Implementadas

* ✅ **Módulo de Usuário:** Cadastro de novos funcionários com senha criptografada (bcrypt).
* ✅ **Autenticação Segura:** Fluxo de login que valida as credenciais e retorna um **JSON Web Token (JWT)**.
* ✅ **Autorização via Middleware:** Endpoints protegidos que só podem ser acessados com um token JWT válido, verificado por um middleware customizado.
* ✅ **Estrutura Profissional:** Organização de projeto seguindo os padrões da comunidade Go para APIs escaláveis.
* ✅ **Containerização:** Configuração pronta para rodar o banco de dados PostgreSQL com Docker.

## 🏛️ Arquitetura

O projeto segue um padrão de **Arquitetura em Camadas** comum em aplicações Go, separando claramente as responsabilidades:

* **/cmd/api:** Ponto de entrada da aplicação, responsável por iniciar o servidor web e carregar as configurações.
* **/config:** Lógica para carregar as variáveis de ambiente (usando Viper).
* **/internal:** Contém o núcleo da lógica da aplicação, inacessível para outros projetos Go (boa prática).
    * **handler:** Camada de apresentação, lida com as requisições HTTP (similar aos Controllers).
    * **service:** Camada de serviço, onde reside a lógica de negócio.
    * **repository:** Camada de acesso a dados, responsável pela comunicação com o banco.
    * **model:** Onde as estruturas (`structs`) de dados do domínio são definidas.
* **/pkg:** Pacotes auxiliares e reutilizáveis, como os serviços de JWT e criptografia de senhas.

## 🚀 Tecnologias Utilizadas

* **Linguagem:** [Go](https://go.dev/)
* **Framework Web / Router:** [Gin](https://github.com/gin-gonic/gin)
* **Banco de Dados:** [PostgreSQL](https://www.postgresql.org/)
* **ORM:** [GORM](https://gorm.io/)
* **Autenticação:** [JWT (golang-jwt)](https://github.com/golang-jwt/jwt)
* **Configuração:** [Viper](https://github.com/spf13/viper) (lendo de arquivos `.env`)
* **Criptografia de Senha:** `bcrypt` (Pacote padrão do Go)
* **Containerização:** [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)

## ⚙️ Como Executar o Projeto

Siga os passos abaixo para configurar e executar a aplicação em seu ambiente local.

### Pré-requisitos

* [Go](https://go.dev/dl/) (versão 1.21 ou superior)
* [Docker](https://www.docker.com/products/docker-desktop/) e Docker Compose
* Um cliente de API como [Postman](https://www.postman.com/) ou [Insomnia](https://insomnia.rest/)

### Passos para Instalação

1.  **Clone o repositório:**
    ```bash
    git clone [https://github.com/](https://github.com/)<SEU_USUARIO>/ponto-api-go.git
    cd ponto-api-go
    ```

2.  **Configure as variáveis de ambiente:**
    Crie um arquivo chamado `.env` na raiz do projeto, copiando o conteúdo do arquivo `.env.example` (que você deve criar), e preencha com suas configurações.
    ```env
    # Porta da API
    API_PORT=8082

    # Configurações do Banco de Dados
    DB_HOST=localhost
    DB_USER=pontouser
    DB_PASSWORD=senha_super_secreta_123
    DB_NAME=ponto_api_db
    DB_PORT=5432

    # Configurações do JWT
    JWT_SECRET_KEY=sua_chave_secreta_super_longa_e_segura_aqui
    JWT_EXPIRATION_IN_HOURS=24
    ```

3.  **Inicie o banco de dados:**
    Com o Docker em execução, inicie o container do PostgreSQL:
    ```bash
    docker-compose up -d
    ```

4.  **Instale as dependências:**
    O Go Modules cuidará disso automaticamente no próximo passo, mas você pode baixar manualmente com:
    ```bash
    go mod tidy
    ```

5.  **Execute a API:**
    ```bash
    go run ./cmd/api/
    ```
    O servidor estará rodando em `http://localhost:8082` (ou na porta que você configurar no `.env`).

## Endpoints da API

A interface da API é a mesma da versão Java.

---

### Módulo de Usuários e Autenticação

#### `POST /api/v1/usuarios`
Cria um novo usuário (funcionário). Rota pública.

* **Body (Exemplo):**
    ```json
    {
        "nome": "Usuário Go",
        "email": "teste.go@email.com",
        "cargo": "Gopher",
        "senha": "senha123"
    }
    ```
* **Resposta de Sucesso (201 Created):**
    ```json
    {
        "id": 1,
        "nome": "Usuário Go",
        "email": "teste.go@email.com",
        "cargo": "Gopher",
        "dataCriacao": "2025-07-31T12:22:03.123Z"
    }
    ```

#### `POST /api/v1/auth/login`
Autentica um usuário e retorna um token JWT. Rota pública.

* **Body (Exemplo):**
    ```json
    {
        "email": "teste.go@email.com",
        "senha": "senha123"
    }
    ```
* **Resposta de Sucesso (200 OK):**
    ```json
    {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0ZS5nb..."
    }
    ```

#### `GET /api/v1/usuarios/me`
Retorna as informações do usuário autenticado. **Rota Protegida.**

* **Authorization Header:**
    * `Authorization: Bearer <SEU_TOKEN_JWT>`
* **Resposta de Sucesso (200 OK):**
    ```json
    {
        "id": 1,
        "nome": "Usuário Go",
        "email": "teste.go@email.com",
        "cargo": "Gopher",
        "dataCriacao": "2025-07-31T12:22:03.123Z"
    }
    ```

---

## 🗺️ Roadmap do Projeto

* [x] Estrutura do projeto em Go.
* [x] Módulo de Usuários (Cadastro).
* [x] Autenticação e Autorização com JWT.
* [ ] Módulo de Ponto (bater o ponto).
* [ ] Consulta de histórico de pontos.
* [ ] Geração de relatórios mensais.
* [ ] Implementar suíte de testes.

## 📄 Licença

Este projeto está sob a licença MIT.
