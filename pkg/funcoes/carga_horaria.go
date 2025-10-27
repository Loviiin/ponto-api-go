package funcoes

import (
	"errors"

	"github.com/Loviiin/ponto-api-go/internal/model"
)

var (
	ErrCargaHorariaAcimaPT   = errors.New("carga horária acima do limite para Part-time/Estagiário (máximo 6h/dia)")
	ErrCargaHorariaAcimaCLT  = errors.New("carga horária acima do limite CLT + horas extras (máximo 10h/dia)")
	ErrCargaHorariaExcessiva = errors.New("carga horária excessiva (máximo 12h/dia)")
)

// ObterCargaHorariaDiaria retorna a carga horária diária em minutos
// com base na hierarquia: Contrato > Cargo > Padrão CLT (8h = 480min)
func ObterCargaHorariaDiaria(contrato *model.Contrato) uint {
	// 1. Prioridade: carga definida no contrato (personalizada)
	if contrato.CargaHorariaDiariaMinutos != nil && *contrato.CargaHorariaDiariaMinutos > 0 {
		return *contrato.CargaHorariaDiariaMinutos
	}

	// 2. Fallback: carga do cargo (padrão do cargo)
	if contrato.Cargo.CargaHorariaDiariaMinutos > 0 {
		return contrato.Cargo.CargaHorariaDiariaMinutos
	}

	// 3. Fallback final: 8h diárias (CLT - Art. 58)
	return 480
}

// ObterCargaHorariaSemanal retorna a carga horária semanal em minutos
// com base na hierarquia: Contrato > Calculado (5 x diária) > 44h CLT
func ObterCargaHorariaSemanal(contrato *model.Contrato) uint {
	// 1. Prioridade: carga semanal definida no contrato
	if contrato.CargaHorariaSemanalMinutos != nil && *contrato.CargaHorariaSemanalMinutos > 0 {
		return *contrato.CargaHorariaSemanalMinutos
	}

	// 2. Calcular com base nos dias trabalhados
	diasSemana := uint(5) // Padrão: segunda a sexta
	if contrato.DiasTrabalhadosSemana != nil && *contrato.DiasTrabalhadosSemana > 0 {
		diasSemana = *contrato.DiasTrabalhadosSemana
	}

	cargaDiaria := ObterCargaHorariaDiaria(contrato)
	return cargaDiaria * diasSemana
}

// ObterDiasTrabalhadosSemana retorna quantos dias por semana o funcionário trabalha
func ObterDiasTrabalhadosSemana(contrato *model.Contrato) uint {
	if contrato.DiasTrabalhadosSemana != nil && *contrato.DiasTrabalhadosSemana > 0 {
		return *contrato.DiasTrabalhadosSemana
	}

	// Padrão: 5 dias (segunda a sexta)
	return 5
}

// ValidarCargaHoraria valida se a carga horária está dentro dos limites legais
func ValidarCargaHoraria(cargaDiariaMinutos uint, tipoContrato string) error {
	// CLT: máximo 8h diárias (480 minutos) + 2h extras (600 minutos)
	// Part-time: máximo 6h diárias (360 minutos)
	// Estagiário: máximo 6h diárias (360 minutos)

	switch tipoContrato {
	case "Part-time", "Estagiário":
		if cargaDiariaMinutos > 360 {
			return ErrCargaHorariaAcimaPT
		}
	case "CLT":
		if cargaDiariaMinutos > 600 {
			return ErrCargaHorariaAcimaCLT
		}
	case "PJ":
		// PJ não tem limite legal, mas vamos estabelecer um máximo razoável
		if cargaDiariaMinutos > 720 { // 12h
			return ErrCargaHorariaExcessiva
		}
	}

	return nil
}
