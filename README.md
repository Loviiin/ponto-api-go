# API de Ponto Eletrônico (Versão em Go)

![Versão da Linguagem](https://img.shields.io/badge/go-1.21+-blue.svg) ![Framework](https://img.shields.io/badge/Gin-v1.9-cyan.svg) ![Licença](https://img.shields.io/badge/license-MIT-green.svg)

## 📖 Sobre o Projeto

A **API de Ponto Eletrônico (Go)** é um sistema backend de alta performance, projetado para gerenciar o registro de jornada de trabalho de funcionários. Este projeto foi desenvolvido como uma peça central de portfólio, demonstrando a aplicação de uma arquitetura limpa, boas práticas e segurança em um ambiente **Go (Golang)**.

Este projeto é a contraparte do [ponto-api (versão em Java/Spring)](<!-- URL_PARA_SEU_REPO_JAVA_AQUI -->), demonstrando a capacidade de resolver o mesmo problema com diferentes stacks de tecnologia, com foco em performance e eficiência de recursos, características marcantes do Go.

## ✨ Features Atuais

* ✅ **CRUD de Usuário:** Cadastro (`POST`) e Leitura (`GET` all, `GET` by ID) de funcionários.
* ✅ **Autenticação Segura:** Fluxo de login (`POST /auth/login`) que valida as credenciais (com senha criptografada via `bcrypt`) e retorna um **JSON Web Token (JWT)**.
* ✅ **Arquitetura Orientada a Domínios:** O projeto foi refatorado para uma estrutura modular, separando as responsabilidades por domínios (`auth`, `usuario`), tornando o sistema mais limpo e escalável.
* ✅ **Containerização do Ambiente:** Configuração pronta para rodar o banco de dados PostgreSQL com `docker-compose`, garantindo um ambiente de desenvolvimento consistente.

## 🏛️ Arquitetura

O projeto segue um padrão de **Arquitetura em Camadas**, com uma organização orientada a domínios para melhor escalabilidade.

* `/cmd/api`: Ponto de entrada da aplicação, responsável por iniciar o servidor e fazer a injeção de dependências.
* `/config`: Lógica para carregar as variáveis de ambiente (usando Viper).
* `/internal`: Contém o núcleo da lógica da aplicação. A estrutura é dividida por domínios:
    * **/auth**: Contém toda a lógica de autenticação.
        * `handler.go`: Lida com as requisições HTTP de login.
        * `service.go`: Orquestra a lógica de negócio da autenticação.
    * **/usuario**: Contém toda a lógica de gerenciamento de usuários.
        * `handler.go`: Lida com as requisições HTTP do CRUD de usuários.
        * `service.go`: Contém as regras de negócio para usuários.
        * `repository.go`: Implementa a comunicação com o banco de dados para a entidade de usuário.
    * **/model**: Onde as estruturas (`structs`) de dados do domínio são definidas.
* `/pkg`: Pacotes auxiliares e reutilizáveis, como os serviços de JWT e criptografia de senhas.

## 🚀 Tecnologias Utilizadas

* **Linguagem:** [Go](https://go.dev/)
* **Framework Web / Router:** [Gin](https://github.com/gin-gonic/gin)
* **Banco de Dados:** [PostgreSQL](https://www.postgresql.org/)
* **ORM:** [GORM](https://gorm.io/)
* **Autenticação:** [JWT (golang-jwt)](https://github.com/golang-jwt/jwt)
* **Configuração:** [Viper](https://github.com/spf13/viper)
* **Criptografia de Senha:** `bcrypt`
* **Containerização:** [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)

## ⚙️ Como Executar o Projeto

Siga os passos abaixo para configurar e executar a aplicação em seu ambiente local.

### Pré-requisitos

* [Go](https://go.dev/dl/) (versão 1.21 ou superior)
* [Docker](https://www.docker.com/products/docker-desktop/) e Docker Compose
* Um cliente de API como [Postman](https://www.postman.com/) ou [Insomnia](https://insomnia.rest/)

### Passos para Instalação

1.  **Clone o repositório:**
    <!-- Lembre-se de substituir <SEU_USUARIO> pelo seu nome de usuário do GitHub -->
    ```bash
    git clone [https://github.com/](https://github.com/)<SEU_USUARIO>/ponto-api-go.git
    cd ponto-api-go
    ```

2.  **Configure as variáveis de ambiente:**
    Copie o arquivo `.env.example` para um novo arquivo chamado `.env`.
    ```bash
    cp .env.example .env
    ```
    Em seguida, revise o arquivo `.env` e preencha com suas configurações, se necessário. Garanta que a `JWT_SECRET_KEY` seja uma string longa e segura.

3.  **Inicie o banco de dados:**
    Com o Docker em execução, inicie o contêiner do PostgreSQL:
    ```bash
    docker-compose up -d
    ```

4.  **Instale as dependências:**
    ```bash
    go mod tidy
    ```

5.  **Execute a API:**
    ```bash
    go run ./cmd/api/main.go
    ```
    O servidor estará rodando em `http://localhost:8082` (ou na porta que você configurar).

## 📖 Endpoints da API

### Autenticação

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

### Usuários

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
        "data_criacao": "2025-08-04T15:30:00.123Z",
        "data_atualizacao": "2025-08-04T15:30:00.123Z"
    }
    ```

#### `GET /api/v1/usuarios`
Retorna uma lista de todos os usuários. *(Atualmente pública, será protegida no futuro)*.

* **Resposta de Sucesso (200 OK):**
    ```json
    [
        {
            "id": 1,
            "nome": "Usuário Go",
            "email": "teste.go@email.com",
            "cargo": "Gopher",
            "data_criacao": "2025-08-04T15:30:00.123Z",
            "data_atualizacao": "2025-08-04T15:30:00.123Z"
        }
    ]
    ```

#### `GET /api/v1/usuarios/{id}`
Retorna as informações de um usuário específico. *(Atualmente pública, será protegida no futuro)*.

* **Resposta de Sucesso (200 OK):**
    ```json
    {
        "id": 1,
        "nome": "Usuário Go",
        "email": "teste.go@email.com",
        "cargo": "Gopher",
        "data_criacao": "2025-08-04T15:30:00.123Z",
        "data_atualizacao": "2025-08-04T15:30:00.123Z"
    }
    ```

---

## 🗺️ Roadmap do Projeto

* [x] Estrutura do projeto por Domínios.
* [x] Módulo de Usuários (Cadastro e Leitura).
* [x] Autenticação com JWT (`/login`).
* [ ] Autorização com Middleware e Rotas Protegidas (`/me`).
* [ ] Módulo de Usuários (Atualização e Deleção).
* [ ] Módulo de Ponto (bater o ponto).
* [ ] Consulta de histórico de pontos.
* [ ] Implementar suíte de testes unitários.
* [ ] Containerizar a aplicação Go com Dockerfile.
* [ ] Adicionar documentação da API com Swagger.

## 📄 Licença

Este projeto está sob a licença MIT.
