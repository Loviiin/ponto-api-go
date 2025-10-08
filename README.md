# Ponto API em Go

<p align="center">
  <img src="https://img.shields.io/badge/go-1.24+-00ADD8?style=for-the-badge&logo=go" alt="Go Version"/>
  <img src="https://img.shields.io/badge/Gin-v1.10-007CDA?style=for-the-badge&logo=gin" alt="Gin Framework"/>
  <img src="https://img.shields.io/badge/PostgreSQL-15-336791?style=for-the-badge&logo=postgresql" alt="PostgreSQL"/>
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker" alt="Docker Ready"/>
  <img src="https://img.shields.io/badge/license-MIT-green?style=for-the-badge" alt="License MIT"/>
</p>

## 📖 Sobre o Projeto

A **Ponto API** é um backend de alta performance para um sistema de Ponto Eletrônico, construído em **Go (Golang)**. Este projeto foi desenhado para ser um exemplo prático de aplicação de arquitetura limpa, boas práticas de desenvolvimento e um robusto sistema de permissões num ambiente moderno e escalável.

O sistema foi projetado com uma **arquitetura multi-tenant**, permitindo que múltiplas empresas utilizem a mesma instância da aplicação de forma segura e com total isolamento de dados.

---

## 🏛️ Conceitos Chave da Arquitetura

Este projeto evoluiu para além de um simples CRUD, adotando um modelo de dados desacoplado e uma arquitetura orientada a domínios.

* **Modelo de Dados Desacoplado:** As responsabilidades foram divididas em quatro entidades principais:
    * `Usuario`: Armazena apenas dados pessoais (identidade).
    * `Empresa`: Armazena apenas dados fiscais (entidade legal).
    * `Localidade`: Representa os locais físicos de trabalho, com endereço e coordenadas para geofence.
    * `Contrato`: Entidade central que representa o vínculo de trabalho, conectando `Usuario`, `Empresa`, `Localidade` e `Cargo`.
* **Arquitetura Orientada a Domínios (DDD-lite):** O código é organizado por áreas de negócio (`usuario`, `empresa`, `contrato`, `ponto`), resultando num sistema modular, com alta coesão e baixo acoplamento.
* **Segurança em Camadas:**
    1.  **Autenticação via JWT:** Garante que apenas utilizadores autenticados acedam a rotas protegidas.
    2.  **Isolamento de Tenant:** A lógica em `repositories` e `services` usa o `empresa_id` (obtido do contrato do utilizador) para garantir que uma empresa nunca aceda aos dados de outra.
    3.  **Autorização Baseada em Permissões (RBAC):** Em vez de cargos fixos (`Admin`), a API usa um sistema granular de `Permissões` (ex: `GERENCIAR_CARGOS`, `AJUSTAR_PONTO_FUNCIONARIOS`) que são atribuídas a `Cargos`. Isto permite uma flexibilidade total na configuração de papéis.
* **Injeção de Dependência:** As dependências são injetadas via construtores, facilitando os testes unitários e o desacoplamento entre as camadas da aplicação.

---

## 🚀 Tecnologias Utilizadas

| Categoria | Tecnologia |
| :--- | :--- |
| **Linguagem** | Go (Golang) |
| **Framework Web** | [Gin](https://github.com/gin-gonic/gin) |
| **Banco de Dados**| [PostgreSQL](https://www.postgresql.org/) |
| **ORM** | [GORM](https://gorm.io/) |
| **Autenticação** | [JWT (golang-jwt)](https://github.com/golang-jwt/jwt) |
| **Configuração** | [Viper](https://github.com/spf13/viper) |
| **Geocodificação**| ViaCEP & OpenCage Geocoder |
| **Containerização**| [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/) |
| **Documentação** | [Swagger](https://swagger.io/) |

---

## ⚙️ Guia de Instalação e Execução

### Pré-requisitos

* Go (versão 1.24 ou superior)
* Docker e Docker Compose
* Cliente de API (Postman, Insomnia, etc.)

### Passos

1.  **Clone o Repositório**
    ```bash
    git clone [https://github.com/Loviiin/ponto-api-go](https://github.com/Loviiin/ponto-api-go)
    cd ponto-api-go
    ```

2.  **Configure as Variáveis de Ambiente**
    Copie o ficheiro de exemplo `.env.example` para `.env` e preencha as variáveis.
    ```bash
    cp .env.example .env
    ```
    *É crucial definir `JWT_SECRET_KEY` e a `API_OPENCAGE`.*

3.  **Inicie os Serviços com Docker Compose**
    Este comando irá construir a imagem da API e iniciar os contentores da aplicação e do banco de dados.
    ```bash
    docker-compose up -d --build
    ```

4.  **Aceda à API**
    O servidor estará a rodar em `http://localhost:8083` (ou na porta configurada). A documentação interativa do Swagger estará disponível em `http://localhost:8083/swagger/index.html`.

---

## 📖 Endpoints da API

O prefixo base para todos os endpoints é `/api/v1`. Endpoints protegidos requerem um `Bearer Token` no cabeçalho `Authorization`.

### 🔑 Autenticação

| Verbo | Endpoint | Descrição | Protegido |
| :--- | :--- | :--- | :--- |
| `POST`| `/auth/signup`| Cria uma nova `Empresa`, `Localidade` (matriz) e o primeiro `Usuario` (com cargo "Dono"). | Não |
| `POST`| `/auth/login` | Autentica um utilizador e retorna um token JWT. | Não |

### 👤 Usuários

| Verbo | Endpoint | Descrição | Protegido | Permissão |
| :--- | :--- | :--- | :--- | :--- |
| `POST`| `/usuarios` | Cria um novo utilizador e o seu `Contrato`. | Sim | `GERENCIAR_CARGOS` (implícito) |
| `GET` | `/usuarios` | Lista todos os utilizadores da empresa. | Sim | N/A |
| `GET` | `/usuarios/me`| Retorna os dados do próprio utilizador logado. | Sim | N/A |
| `GET` | `/usuarios/{id}`| Busca um utilizador específico por ID. | Sim | N/A |
| `PUT` | `/usuarios/{id}`| Atualiza dados de um utilizador. | Sim | `EDITAR_USUARIO` ou `EDITAR_PROPRIA_CONTA` |
| `DELETE`| `/usuarios/{id}`| Apaga um utilizador e o seu contrato. | Sim | `DELETAR_USUARIO` ou `DELETAR_PROPRIA_CONTA` |

### 🏢 Empresas, Localidades e Cargos

| Verbo | Endpoint | Descrição | Protegido | Permissão |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/empresas` | Lista todas as empresas (endpoint de admin/super-app). | Sim | (a definir) |
| `GET` | `/empresas/{id}`| Busca uma empresa por ID. | Sim | N/A |
| `PUT` | `/empresas/{id}`| Atualiza os dados da empresa. | Sim | `EDITAR_EMPRESA` |
| `POST`| `/localidades`| Cria uma nova localidade (filial) para a empresa. | Sim | `GERENCIAR_LOCALIDADES` |
| `GET` | `/empresas/{id}/localidades`| Lista as localidades de uma empresa. | Sim | `GERENCIAR_LOCALIDADES` |
| `GET` | `/cargos` | Lista todos os cargos da empresa. | Sim | N/A |
| `PUT` | `/cargos/{id}` | Atualiza um cargo. | Sim | `GERENCIAR_CARGOS` |
| `POST`| `/cargos/{id}/permissoes/{pId}`| Associa uma permissão a um cargo. | Sim | `GERENCIAR_CARGOS` |

### 🕒 Ponto e Banco de Horas

| Verbo | Endpoint | Descrição | Protegido | Permissão |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/pontos` | Registra uma batida de ponto (com geofence). | Sim | N/A |
| `GET` | `/pontos/meus-registros`| Lista os pontos do utilizador logado. | Sim | N/A |
| `GET` | `/pontos/usuario/{id}`| Lista os pontos de um funcionário específico. | Sim | `VISUALIZAR_PONTO_FUNCIONARIOS` |
| `POST` | `/pontos/ajuste` | Admin adiciona um registo de ponto manual. | Sim | `AJUSTAR_PONTO_FUNCIONARIOS` |
| `GET` | `/bancohoras/saldo/usuario/{id}`| Consulta o saldo de horas de um funcionário. | Sim | `VER_SALDO_FUNCIONARIOS` (se não for o próprio) |
| `POST` | `/bancohoras/fechamento/usuario/{id}`| Admin força o fechamento do dia para um funcionário. | Sim | `EDITAR_SALDO_FUNCIONARIOS` |

*(Endpoints de Justificativas e Permissões foram omitidos por brevidade)*

### 📤 Exportação de Relatórios de Ponto (Novo)

Dois endpoints permitem exportar registros de ponto em CSV ou PDF para um intervalo de datas.

| Verbo | Endpoint | Descrição | Protegido | Permissão |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/relatorios/ponto/meus-registros/export` | Exporta os próprios registros de ponto. | Sim | N/A |
| `GET` | `/relatorios/ponto/usuario/{id}/export` | Exporta registros de um funcionário específico. | Sim | `VISUALIZAR_PONTO_FUNCIONARIOS` |

Query Params obrigatórios:
| Nome | Descrição | Formato | Exemplo |
| :--- | :--- | :--- | :--- |
| `data_inicio` | Data inicial (inclusiva) | AAAA-MM-DD | `2025-10-01` |
| `data_fim` | Data final (inclusiva) | AAAA-MM-DD | `2025-10-31` |
| `formato` | Formato de exportação | `csv` ou `pdf` | `csv` |

Exemplo de chamada:
```
GET /api/v1/relatorios/ponto/meus-registros/export?data_inicio=2025-10-01&data_fim=2025-10-31&formato=csv
Authorization: Bearer <TOKEN>
```

Headers de resposta:
```
Content-Type: text/csv (ou application/pdf)
Content-Disposition: attachment; filename="relatorio_ponto_<user>_<inicio>_<fim>.<ext>"
```

Notas sobre o PDF: layout inclui cabeçalho centralizado, metadados (período e data de geração), paginação no rodapé, linhas alternadas e total de registros.

### 🪞 Espelho de Ponto (Novo)

O espelho de ponto consolida as marcações de entrada/saída por dia, calcula tempo trabalhado, horas previstas e saldo acumulado no período.

| Verbo | Endpoint | Descrição | Protegido | Permissão |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/relatorios/ponto/espelho/me` | Gera o espelho do utilizador logado. | Sim | N/A |
| `GET` | `/relatorios/ponto/espelho/usuario/{id}` | Gera o espelho de um funcionário. | Sim | `VISUALIZAR_PONTO_FUNCIONARIOS` |

Query Params:
| Nome | Descrição | Formato | Exemplo |
| :--- | :--- | :--- | :--- |
| `data_inicio` | Data inicial (inclusiva) | AAAA-MM-DD | `2025-10-01` |
| `data_fim` | Data final (inclusiva) | AAAA-MM-DD | `2025-10-07` |

Response (exemplo simplificado):
```json
{
    "usuario_id": 42,
    "empresa_id": 7,
    "inicio": "2025-10-01",
    "fim": "2025-10-07",
    "total_trabalhado_minutos": 2280,
    "total_previsto_minutos": 2400,
    "saldo_acumulado_minutos": -120,
    "dias": [
        {
            "data": "2025-10-01",
            "registros": [
                {"id": 10, "timestamp": "2025-10-01T08:00:00Z"},
                {"id": 11, "timestamp": "2025-10-01T12:00:00Z"},
                {"id": 12, "timestamp": "2025-10-01T13:00:00Z"},
                {"id": 13, "timestamp": "2025-10-01T17:00:00Z"}
            ],
            "total_trabalhado_minutos": 480,
            "horas_previstas_minutos": 480,
            "saldo_dia_minutos": 0,
            "fechado": false,
            "inconsistente": false
        }
    ]
}
```

Regras atuais de cálculo:
* Registros são agrupados por dia (timezone UTC no momento).
* Marcações pares são consideradas pares Entrada/Saída sequenciais; marcação ímpar → dia marcado como `inconsistente` e a última sobra é ignorada no cálculo.
* Pausas (almoço) são inferidas pelos pares; não há validação de sobreposição.
* Carga horária prevista diária: obtida do contrato do utilizador; se ausente, assume 480 minutos (8h) temporariamente.
* Intervalos invertidos (data_inicio > data_fim) são normalizados automaticamente.
* Campo `fechado` ainda é placeholder (integração futura com logs de fechamento de banco de horas).

Melhorias Futuras Planeadas:
* Usar timezone configurável por localidade.
* Marcar dia como `fechado` com base em `LogBancoHoras`.
* Exportar espelho em PDF/CSV.
* Mostrar saldo acumulado também em formato HH:MM.

---

## 🗺️ Próximos Passos (Roadmap)

A fundação do sistema está robusta e pronta para escalar. Os próximos passos focam em enriquecer as funcionalidades de gestão:

-   [ ] **Épico: Gestão de Contratos:** Criar endpoints para `PUT`, `GET` e `DELETE` de contratos, permitindo transferir um funcionário de cargo ou localidade.
-   [ ] **Épico: Relatórios:** Desenvolver endpoints que gerem relatórios de folha de ponto (`espelho de ponto`) por funcionário e por período.
-   [ ] **Testes Unitários e de Integração:** Aumentar a cobertura de testes, especialmente para os serviços de `auth`, `usuario` e `bancohoras`, para garantir a estabilidade após a refatoração.
-   [ ] **Melhorar o `Scheduler`:** Tornar a tarefa de fechamento diário mais resiliente e com melhores logs.
-   [ ] **Validações Avançadas:** Implementar validações mais complexas (ex: garantir que um CNPJ ou CPF é matematicamente válido).
-   [ ] **Upload de Documentos:** Adicionar a funcionalidade de upload para comprovativos de morada ou atestados, ligada ao módulo de `Justificativas`.

## 📄 Licença

Este projeto está sob a licença MIT.
