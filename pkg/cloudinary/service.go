package cloudinary

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Service interface para operações com Cloudinary
type Service interface {
	UploadAvatar(file multipart.File, userID uint, filename string) (string, error)
	DeleteAvatar(publicID string) error
}

type service struct {
	cld *cloudinary.Cloudinary
}

// NewService cria uma nova instância do serviço Cloudinary
// cloudinaryURL deve estar no formato: cloudinary://API_KEY:API_SECRET@CLOUD_NAME
func NewService(cloudinaryURL string) (Service, error) {
	if cloudinaryURL == "" {
		return nil, errors.New("CLOUDINARY_URL não configurada")
	}

	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar Cloudinary: %w", err)
	}

	return &service{cld: cld}, nil
}

// UploadAvatar faz upload de um avatar para o Cloudinary
// Retorna a URL segura (HTTPS) da imagem
func (s *service) UploadAvatar(file multipart.File, userID uint, filename string) (string, error) {
	start := time.Now() // Medição de performance
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Extrair extensão do arquivo
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg" // padrão
	}

	// Criar public_id único: avatars/user_123_1234567890
	publicID := fmt.Sprintf("avatars/user_%d_%d", userID, time.Now().Unix())

	// Helper para ponteiro bool
	boolPtr := func(b bool) *bool { return &b }

	// Configurar parâmetros do upload
	uploadParams := uploader.UploadParams{
		PublicID:       publicID,
		Folder:         "ponto-api/avatars", // Organizar em pasta
		ResourceType:   "image",
		Transformation: "c_fill,g_face,h_400,w_400",  // Crop automático em 400x400 focando no rosto
		Format:         strings.TrimPrefix(ext, "."), // jpg, png
		Overwrite:      boolPtr(false),               // Não sobrescrever se existir
		UniqueFilename: boolPtr(true),                // Garantir nome único
		Tags:           []string{"avatar", "user"},   // Tags para organização
	}

	// Fazer upload
	result, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer upload no Cloudinary: %w", err)
	}

	// Log de performance (remover em produção)
	elapsed := time.Since(start)
	fmt.Printf("[Cloudinary] Upload concluído em %v (userID: %d)\n", elapsed, userID)

	// Retornar URL segura (HTTPS)
	return result.SecureURL, nil
}

// DeleteAvatar deleta um avatar do Cloudinary pelo public_id
// publicID exemplo: "ponto-api/avatars/user_123_1234567890"
func (s *service) DeleteAvatar(publicID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Extrair o public_id da URL se for uma URL completa
	// Se for URL: https://res.cloudinary.com/cloud/image/upload/v123/ponto-api/avatars/user_1_123.jpg
	// Precisamos extrair: ponto-api/avatars/user_1_123
	if strings.Contains(publicID, "cloudinary.com") {
		parts := strings.Split(publicID, "/upload/")
		if len(parts) >= 2 {
			// Pegar tudo após "/upload/v123/"
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) > 1 {
				// Remover versão (v123) e reconstruir path
				publicID = strings.Join(pathParts[1:], "/")
				// Remover extensão (.jpg, .png)
				publicID = strings.TrimSuffix(publicID, filepath.Ext(publicID))
			}
		}
	}

	// Se já é apenas o public_id, remover extensão se houver
	publicID = strings.TrimSuffix(publicID, filepath.Ext(publicID))

	// Helper para ponteiro bool
	boolPtr := func(b bool) *bool { return &b }

	deleteParams := uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
		Invalidate:   boolPtr(true), // Invalidar CDN cache
	}

	_, err := s.cld.Upload.Destroy(ctx, deleteParams)
	if err != nil {
		return fmt.Errorf("erro ao deletar avatar do Cloudinary: %w", err)
	}

	return nil
}

// ExtractPublicIDFromURL extrai o public_id de uma URL do Cloudinary
// Útil para deletar imagens antigas quando o usuário faz novo upload
func ExtractPublicIDFromURL(url string) string {
	if !strings.Contains(url, "cloudinary.com") {
		return ""
	}

	parts := strings.Split(url, "/upload/")
	if len(parts) < 2 {
		return ""
	}

	// Pegar tudo após "/upload/v123/"
	pathParts := strings.Split(parts[1], "/")
	if len(pathParts) <= 1 {
		return ""
	}

	// Remover versão (v123) e reconstruir path
	publicID := strings.Join(pathParts[1:], "/")

	// Remover extensão
	publicID = strings.TrimSuffix(publicID, filepath.Ext(publicID))

	return publicID
}
