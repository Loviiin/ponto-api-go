package empresa

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
)

type EmpresaService interface {
	CreateEmpresa(empresa *model.Empresa) error
	GetAllEmpresasSer() ([]model.Empresa, error)
	GetEmpresaByIDSer(idempresa uint) (*model.Empresa, error)
	UpdateEmpresaSer(idempresa uint, dados map[string]interface{}) error
	DeleteEmpresaSer(idempresa uint) error
}

type empresaService struct {
	empresaRepo EmpresaRepository
}

func NewEmpresaService(repo EmpresaRepository) EmpresaService {
	return &empresaService{
		empresaRepo: repo,
	}
}

func (s *empresaService) CreateEmpresa(empresa *model.Empresa) error {
	return s.empresaRepo.CreateEmpresa(empresa)
}

func (s *empresaService) GetAllEmpresasSer() ([]model.Empresa, error) {
	return s.empresaRepo.GetAllEmpresas()
}

func (s *empresaService) GetEmpresaByIDSer(idempresa uint) (*model.Empresa, error) {
	return s.empresaRepo.GetEmpresaByID(idempresa)
}

func (s *empresaService) UpdateEmpresaSer(idempresa uint, dados map[string]interface{}) error {
	_, err := s.empresaRepo.FindByID(idempresa)
	if err != nil {
		return err
	}
	return s.empresaRepo.UpdateEmpresa(idempresa, dados)
}

func (s *empresaService) DeleteEmpresaSer(idempresa uint) error {
	_, err := s.empresaRepo.FindByID(idempresa)
	if err != nil {
		return err
	}

	return s.empresaRepo.DeleteEmpresa(idempresa)
}
