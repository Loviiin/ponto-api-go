# Ponto API em Go

<p align="center">
  <img src="https://img.shields.io/badge/go-1.24+-00ADD8?style=for-the-badge&logo=go" alt="Go Version"/>
  <img src="https://img.shields.io/badge/Gin-v1.10-007CDA?style=for-the-badge&logo=gin" alt="Gin Framework"/>
  <img src="https://img.shields.io/badge/PostgreSQL-15-336791?style=for-the-badge&logo=postgresql" alt="PostgreSQL"/>
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker" alt="Docker Ready"/>
  <img src="https://img.shields.io/badge/license-AGPLv3%20%7C%20Commercial-blue?style=for-the-badge" alt="License"/>
</p>

## 📖 Sobre o Projeto

A **Ponto API** é um backend de alta performance para um sistema de Ponto Eletrônico, construído em **Go (Golang)**. Este projeto foi desenhado não apenas para ser funcional, mas também para servir como um exemplo prático de aplicação de arquitetura limpa, boas práticas de desenvolvimento e segurança em um ambiente moderno.

O sistema foi projetado desde o início com uma **arquitetura multi-tenant**, permitindo que múltiplas empresas utilizem a mesma instância da aplicação de forma segura e isolada.



## 🏛️ Conceitos Chave da Arquitetura

Este projeto não é apenas um CRUD. Ele foi construído sobre uma fundação de princípios de software robustos:

* **Arquitetura Orientada a Domínios:** Inspirado no (DDD), o código é organizado por áreas de negócio (`usuario`, `empresa`, `cargo`, `ponto`). Isso resulta em um sistema modular, com alta coesão e baixo acoplamento, facilitando a manutenção e a escalabilidade.
* **Multi-Tenancy:** O sistema utiliza um modelo de banco de dados compartilhado com `empresa_id` em todas as entidades relevantes, garantindo que os dados de uma empresa sejam completamente isolados dos de outra.
* **Segurança em Camadas:** A segurança é aplicada em múltiplos níveis:
    1.  **Autenticação via JWT:** Garante que apenas usuários logados acessem a maioria dos recursos.
    2.  **Isolamento de Tenant:** A lógica em `repositories` e `services` garante que um usuário só possa ver e modificar dados da sua própria empresa.
    3.  **Autorização Baseada em Cargos (RBAC):** Um `RoleAuthMiddleware` protege endpoints críticos, garantindo que apenas usuários com cargos específicos (ex: `ADMIN`) possam realizar operações sensíveis, como editar dados da empresa.
* **Injeção de Dependência:** As dependências (como repositórios e serviços) são injetadas via construtores, facilitando os testes unitários e o desacoplamento entre as camadas.

---

## 🚀 Tecnologias Utilizadas

| Categoria         | Tecnologia                                                                                             |
| :---------------- | :----------------------------------------------------------------------------------------------------- |
| **Linguagem** | Go (Golang)                                                                                            |
| **Framework Web** | [Gin](https://github.com/gin-gonic/gin)                                                                |
| **Banco de Dados** | [PostgreSQL](https://www.postgresql.org/)                                                              |
| **ORM** | [GORM](https://gorm.io/)                                                                               |
| **Autenticação** | [JWT (golang-jwt)](https://github.com/golang-jwt/jwt)                                                    |
| **Configuração** | [Viper](https://github.com/spf13/viper)                                                                |
| **Containerização** | [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)                 |

---

## ⚙️ Guia de Instalação e Execução

Siga os passos abaixo para ter o ambiente completo rodando localmente.

### Pré-requisitos

* Go (versão 1.24 ou superior)
* Docker e Docker Compose
* Um cliente de API como [Postman](https://www.postman.com/) ou [Insomnia](https://insomnia.rest/)

### Passos

1.  **Clone o Repositório**
    ```bash
    git clone https://github.com/Loviiin/ponto-api-go
    cd ponto-api-go
    ```

2.  **Configure as Variáveis de Ambiente**
    Copie o arquivo de exemplo e, se necessário, ajuste as variáveis.
    ```bash
    cp .env.example .env
    ```
    *É crucial definir uma `JWT_SECRET_KEY` forte e segura.*

3.  **Inicie o Banco de Dados com Docker**
    Este comando irá baixar a imagem do PostgreSQL e iniciar o contêiner em segundo plano.
    ```bash
    docker-compose up -d
    ```

4.  **Instale as Dependências do Go**
    ```bash
    go mod tidy
    ```

5.  **Execute a API**
    ```bash
    go run ./cmd/api/main.go
    ```
    O servidor estará rodando em `http://localhost:8083` (ou na porta configurada no seu `.env`).

---

## API Endpoints

O prefixo base para todos os endpoints é `/api/v1`.

### 🔑 Autenticação

| Verbo  | Endpoint       | Descrição                                    | Protegido |
| :----- | :------------- | :------------------------------------------- | :-------- |
| `POST` | `/auth/login`  | Autentica um usuário e retorna um token JWT. | Não       |

### 🏢 Empresas

| Verbo    | Endpoint         | Descrição                                 | Protegido | Permissão Extra |
| :------- | :--------------- | :---------------------------------------- |:----------| :-------------- |
| `POST`   | `/empresas`      | Cria uma nova empresa.                    | Sim       |                 |
| `GET`    | `/empresas`      | Lista todas as empresas.                  | Sim       |                 |
| `GET`    | `/empresas/{id}` | Busca uma empresa por ID.                 | Não       |                 |
| `PUT`    | `/empresas/{id}` | Atualiza os dados da própria empresa.     | Sim       | Cargo: `ADMIN`  |
| `DELETE` | `/empresas/{id}` | Deleta a própria empresa.                 | Sim       | Cargo: `ADMIN`  |

### 👤 Usuários

| Verbo    | Endpoint         | Descrição                                     | Protegido |
| :------- | :--------------- | :-------------------------------------------- | :-------- |
| `POST`   | `/usuarios`      | Cria um novo usuário (funcionário).           | Não       |
| `GET`    | `/usuarios`      | Lista os usuários da empresa do requisitante. | Sim       |
| `GET`    | `/usuarios/me`   | Retorna os dados do próprio usuário logado.   | Sim       |
| `PUT`    | `/usuarios/{id}` | Atualiza os dados do próprio usuário.         | Sim       |
| `DELETE` | `/usuarios/{id}` | Deleta o próprio usuário.                     | Sim       |

### 🗂️ Cargos

| Verbo    | Endpoint       | Descrição                                 | Protegido |
| :------- | :------------- | :---------------------------------------- | :-------- |
| `POST`   | `/cargos`      | Cria um novo cargo para a empresa.        | Sim       |
| `GET`    | `/cargos`      | Lista os cargos da empresa.               | Sim       |
| `PUT`    | `/cargos/{id}` | Atualiza um cargo da empresa.             | Sim       |
| `DELETE` | `/cargos/{id}` | Deleta um cargo da empresa.               | Sim       |

### 🕒 Ponto

| Verbo  | Endpoint  | Descrição                                     | Protegido |
| :----- | :-------- | :-------------------------------------------- | :-------- |
| `POST` | `/pontos` | Registra uma batida de ponto (entrada/saída). | Sim       |

---

## 📜 Licenciamento

Este projeto opera sob um modelo de **dual-license**, oferecendo flexibilidade para diferentes tipos de uso:

1.  **Community Edition (Gratuita):**
    *   **Licença:** [GNU AGPLv3](LICENSE)
    *   **Ideal para:** Estudantes, startups em estágio inicial e projetos de código aberto.
    *   **Descrição:** Uma versão funcional e básica, perfeita para aprender e para uso em projetos que também são de código aberto. Requer que quaisquer modificações distribuídas ou usadas em um serviço de rede também sejam de código aberto.

2.  **Enterprise Edition (Comercial):**
    *   **Licença:** Comercial
    *   **Ideal para:** Empresas que necessitam de funcionalidades avançadas, suporte prioritário e a flexibilidade de uma licença comercial.
    *   **Descrição:** Inclui todos os recursos da versão community, além de funcionalidades exclusivas, como integrações avançadas, relatórios personalizados, suporte técnico dedicado e a permissão para manter o código-fonte modificado como proprietário. Para adquirir uma licença comercial, entre em contato.

---
