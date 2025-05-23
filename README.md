```markdown
# Projeto POC - Gerenciamento de Upload de Arquivos

Este projeto é uma **Prova de Conceito (POC)** para demonstrar a implementação de um sistema de upload e gerenciamento de arquivos utilizando **Golang** e **GORM** como ORM para interação com o banco de dados SQLite.

## Funcionalidades

- **Upload de Arquivos**: Permite o envio de arquivos para o servidor.
- **Validação de Arquivos**: Verifica se o arquivo já existe no banco de dados antes de realizar o upload.
- **Armazenamento de Metadados**: Salva informações como nome, tipo, tamanho e caminho do arquivo no banco de dados.
- **Migração Automática**: Criação automática da tabela `assets` no banco de dados utilizando o GORM.

## Estrutura do Projeto

- **`internal/models/asset.go`**: Define o modelo `Asset` com os campos necessários para armazenar os metadados dos arquivos.
- **`internal/handlers/upload-file.go`**: Contém o handler responsável por gerenciar o upload de arquivos.
- **`internal/database`**: Configuração e inicialização do banco de dados SQLite.
- **`Dockerfile`**: Configuração para criar uma imagem Docker do projeto, incluindo a compilação do binário Go e dependências adicionais.

## Tecnologias Utilizadas

- **Linguagem**: Go
- **Framework Web**: Gin
- **Banco de Dados**: SQLite
- **ORM**: GORM
- **Docker**: Para containerização da aplicação

## Como Executar

### Pré-requisitos

- **Go** (versão 1.23.2 ou superior)
- **Docker** (opcional, para execução em container)

### Passos

1. Clone o repositório:
   ```bash
   git clone <url-do-repositorio>
   cd <nome-do-repositorio>
   ```

2. Instale as dependências:
   ```bash
   go mod tidy
   ```

3. Execute a aplicação:
   ```bash
   go run cmd/main.go
   ```

4. Para executar via Docker:
   ```bash
   docker build -t poc-upload .
   docker run -p 8080:8080 poc-upload
   ```

### Endpoints

- **POST** `/upload`: Endpoint para realizar o upload de arquivos.  
  **Exemplo de resposta**:
  ```json
  {
    "message": "Arquivo salvo com sucesso",
    "filename": "exemplo.pdf",
    "path": "temp/exemplo.pdf",
    "size": 12345,
    "type": "application/pdf"
  }
  ```

## Estrutura do Banco de Dados

A tabela `assets` possui os seguintes campos:

- `id`: Identificador único.
- `name`: Nome do arquivo.
- `type`: Tipo MIME do arquivo.
- `size`: Tamanho do arquivo em bytes.
- `path`: Caminho onde o arquivo foi salvo.
- `encrypted`: Indica se o arquivo está criptografado.
- `created_at`, `updated_at`, `deleted_at`: Campos gerenciados automaticamente pelo GORM.

## Observações

- Arquivos enviados são armazenados no diretório `temp/`.
- Os containers Docker têm limites de recursos configurados:
  - **Memória total**: Máximo de 4GB distribuídos entre os serviços
  - **App**: 2GB de memória e 1.5 CPUs
  - **FFmpeg**: 1.5GB de memória e 2.0 CPUs
  - **Redis**: 500MB de memória e 0.5 CPUs
