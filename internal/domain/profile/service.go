package profile

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/audit"
	"github.com/Loviiin/ponto-api-go/pkg/cache"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"github.com/Loviiin/ponto-api-go/pkg/validator"
	"gorm.io/gorm"
)

// Service define a interface para operações de perfil
type Service interface {
	// Perfil
	GetMyProfile(userID uint) (*ProfileResponse, error)
	UpdateProfile(userID uint, req UpdateProfileRequest, ip, userAgent string) (*ProfileResponse, error)
	ChangePassword(userID uint, req ChangePasswordRequest, ip, userAgent string) error
	UpdateAvatar(userID uint, avatarURL string) error
	UpdateCPF(adminID, targetUserID uint, req UpdateCPFRequest, ip, userAgent string) error

	// Estatísticas
	GetMyStats(userID, empresaID uint) (*StatsResponse, error)
	GetCalendar(userID uint, month, year int) (*CalendarioResponse, error)

	// Permissões
	GetMyPermissions(userID uint) (*PermissoesResponse, error)

	// Atividades Recentes
	GetRecentActivity(userID uint, limit int) (*RecentActivityResponse, error)
}

type service struct {
	repo        Repository
	db          *gorm.DB
	cache       cache.Service
	auditLogger audit.Service
}

// NewService cria uma nova instância do service
func NewService(repo Repository, db *gorm.DB, cache cache.Service, auditLogger audit.Service) Service {
	return &service{
		repo:        repo,
		db:          db,
		cache:       cache,
		auditLogger: auditLogger,
	}
}

// GetMyProfile retorna o perfil completo do usuário
func (s *service) GetMyProfile(userID uint) (*ProfileResponse, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, err
	}

	profile := &ProfileResponse{
		ID:        user.ID,
		Nome:      user.Nome,
		Email:     user.Email,
		CPF:       maskCPF(user.CPF),
		Telefone:  user.Telefone,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}

	// Informações da empresa
	if user.Contrato.ID > 0 && user.Contrato.Empresa.ID > 0 {
		profile.Empresa = EmpresaInfo{
			ID:          user.Contrato.Empresa.ID,
			Nome:        user.Contrato.Empresa.NomeFantasia,
			CNPJ:        maskCNPJ(user.Contrato.Empresa.CNPJ),
			RazaoSocial: user.Contrato.Empresa.RazaoSocial,
		}

		// Informações do cargo
		if user.Contrato.Cargo.ID > 0 {
			profile.Cargo = &CargoInfo{
				ID:   user.Contrato.Cargo.ID,
				Nome: user.Contrato.Cargo.Nome,
			}
		}

		// Informações do contrato
		profile.Contrato = &ContratoInfo{
			ID:                  user.Contrato.ID,
			DataInicio:          user.Contrato.DataAdmissao,
			DataFim:             user.Contrato.DataDemissao,
			TipoContrato:        user.Contrato.TipoContrato,
			CargaHorariaSemanal: calculateCargaHorariaSemanal(user.Contrato),
			CargaHorariaMensal:  calculateCargaHorariaMensal(user.Contrato),
		}

		// Adicionar horários esperados do cargo
		if user.Contrato.Cargo.ID > 0 {
			profile.Contrato.HorarioEntrada = minutesToTimeString(user.Contrato.Cargo.EntradaEsperadaMinutos)
			profile.Contrato.HorarioSaida = minutesToTimeString(user.Contrato.Cargo.SaidaEsperadaMinutos)
			profile.Contrato.IntervaloMinutos = int(user.Contrato.Cargo.MinutosAlmocoEsperado)
		}

		// Buscar localidades da empresa
		localidades, err := s.repo.GetLocalidadesByEmpresa(user.Contrato.EmpresaID)
		if err == nil && len(localidades) > 0 {
			profile.Localidades = make([]LocalidadeInfo, len(localidades))
			for i, loc := range localidades {
				endereco := fmt.Sprintf("%s, %s - %s, %s/%s",
					loc.Logradouro, loc.Numero, loc.Bairro, loc.Cidade, loc.Estado)
				profile.Localidades[i] = LocalidadeInfo{
					ID:        loc.ID,
					Nome:      loc.Nome,
					Endereco:  endereco,
					Latitude:  loc.Latitude,
					Longitude: loc.Longitude,
					Raio:      int(loc.RaioGeofenceMetros),
				}
			}
		}
	}

	// Buscar permissões
	permissoes, err := s.repo.GetPermissoesByUserID(userID)
	if err == nil && len(permissoes) > 0 {
		profile.Permissoes = make([]PermissaoInfo, len(permissoes))
		for i, perm := range permissoes {
			profile.Permissoes[i] = PermissaoInfo{
				ID:        perm.ID,
				Nome:      perm.Nome,
				Descricao: perm.Descricao,
			}
		}
	}

	return profile, nil
}

// UpdateProfile atualiza os dados editáveis do perfil
// IMPORTANTE: Este método usa PATCH semântico - apenas atualiza campos enviados
// A senha NUNCA é tocada aqui (tem endpoint separado)
func (s *service) UpdateProfile(userID uint, req UpdateProfileRequest, ip, userAgent string) (*ProfileResponse, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, err
	}

	// Guardar dados antigos para audit log
	dadosAntigos := map[string]interface{}{
		"nome":     user.Nome,
		"email":    user.Email,
		"telefone": user.Telefone,
	}

	// CRITICAL FIX: Usar map para UPDATE SELETIVO - apenas campos enviados
	// Isso evita sobrescrever senha e outros campos não incluídos
	updates := make(map[string]interface{})

	if req.Nome != nil && *req.Nome != "" {
		updates["nome"] = *req.Nome
	}

	if req.Telefone != nil {
		// DATA VALIDATION: Sanitizar e validar telefone brasileiro
		telefoneSanitizado := validator.SanitizarTelefone(*req.Telefone)
		if telefoneSanitizado != "" {
			if err := validator.ValidarTelefoneBrasileiro(*req.Telefone); err != nil {
				return nil, err
			}
			updates["telefone"] = telefoneSanitizado
		} else {
			// Permitir limpar o telefone
			updates["telefone"] = ""
		}
	}

	if req.Email != nil && *req.Email != "" {
		// DATA VALIDATION: Normalizar email
		novoEmail := password.NormalizarEmail(*req.Email)

		// Verificar se email já está em uso
		err = s.db.Where("email = ? AND id != ?", novoEmail, userID).First(&model.Usuario{}).Error
		if err == nil {
			return nil, errors.New("email já está em uso por outro usuário")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("erro ao verificar email")
		}

		updates["email"] = novoEmail
	}

	// CRITICAL: Usar db.Model().Updates() para UPDATE PARCIAL
	// Isso garante que APENAS os campos no map sejam atualizados
	// A senha e outros campos NÃO serão tocados
	if len(updates) > 0 {
		// DEBUG LOG: Ver exatamente o que será atualizado
		fmt.Printf("[DEBUG] UpdateProfile - userID: %d, updates: %+v\n", userID, updates)

		err = s.db.Model(&model.Usuario{}).Where("id = ?", userID).Updates(updates).Error
		if err != nil {
			return nil, errors.New("erro ao atualizar perfil")
		}

		// Invalidar cache após update bem-sucedido
		s.repo.InvalidateUserCache(userID)
	}

	// Obter empresa_id do contrato para audit log
	var empresaID uint
	if user.Contrato.ID > 0 {
		empresaID = user.Contrato.EmpresaID
	}

	// AUDIT LOG: Registrar alteração de perfil
	if len(updates) > 0 && s.auditLogger != nil {
		_ = s.auditLogger.LogAction(&userID, &empresaID, "UPDATE_PROFILE", "usuario", userID, dadosAntigos, updates, ip, userAgent)
	}

	// Retornar perfil atualizado
	return s.GetMyProfile(userID)
}

// ChangePassword altera a senha do usuário
func (s *service) ChangePassword(userID uint, req ChangePasswordRequest, ip, userAgent string) error {
	// Validar se nova senha e confirmação são iguais
	if req.NovaSenha != req.ConfirmarSenha {
		return errors.New("nova senha e confirmação não coincidem")
	}

	// SECURITY FIX: Validar força da nova senha
	if err := password.ValidarForcaSenha(req.NovaSenha); err != nil {
		return err
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("usuário não encontrado")
		}
		return err
	}

	// Verificar senha atual
	if !password.VerificaHashSenha(req.SenhaAtual, user.Senha) {
		return errors.New("senha atual incorreta")
	}

	// Hash da nova senha
	hashedPassword, err := password.CriptografaSenha(req.NovaSenha)
	if err != nil {
		return errors.New("erro ao processar nova senha")
	}

	user.Senha = hashedPassword

	// Salvar alterações
	if err := s.repo.UpdateUser(user); err != nil {
		return errors.New("erro ao atualizar senha")
	}

	// AUDIT LOG: Registrar alteração de senha
	var empresaID uint
	if user.Contrato.ID > 0 {
		empresaID = user.Contrato.EmpresaID
	}

	if s.auditLogger != nil {
		_ = s.auditLogger.LogAction(&userID, &empresaID, "CHANGE_PASSWORD", "usuario", userID, nil, map[string]interface{}{"changed": true}, ip, userAgent)
	}

	return nil
}

// UpdateAvatar atualiza a URL do avatar do usuário
func (s *service) UpdateAvatar(userID uint, avatarURL string) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("usuário não encontrado")
		}
		return err
	}

	user.Avatar = avatarURL

	if err := s.repo.UpdateUser(user); err != nil {
		return errors.New("erro ao atualizar avatar")
	}

	return nil
}

// GetMyStats retorna estatísticas de ponto do usuário
func (s *service) GetMyStats(userID, empresaID uint) (*StatsResponse, error) {
	now := time.Now()
	currentMonth := int(now.Month())
	currentYear := now.Year()

	// Banco de horas
	bancoHoras, err := s.repo.GetLatestBancoHoras(userID, empresaID)
	var bancoInfo BancoHorasInfo
	if err == nil && bancoHoras != nil {
		bancoInfo = BancoHorasInfo{
			SaldoAtual:        bancoHoras.SaldoNovoMinutos,
			SaldoFormatado:    formatMinutesToHours(bancoHoras.SaldoNovoMinutos),
			UltimaAtualizacao: bancoHoras.Data,
		}
	} else {
		bancoInfo = BancoHorasInfo{
			SaldoAtual:        0,
			SaldoFormatado:    "00:00",
			UltimaAtualizacao: now,
		}
	}

	// Estatísticas do mês atual
	monthStats, err := s.repo.GetPontoStatsByUserAndMonth(userID, currentMonth, currentYear)
	if err != nil {
		monthStats = &PontoMonthStats{}
	}

	mesAtual := MesAtualStats{
		Mes:                   currentMonth,
		Ano:                   currentYear,
		DiasTrabalados:        monthStats.DiasTrabalados,
		DiasAusentes:          0, // Calcular com base em dias úteis
		HorasTrabalhadas:      monthStats.MinutosTrabalhados,
		TotalAtrasos:          monthStats.TotalAtrasos,
		TotalSaidasAdiantadas: monthStats.TotalSaidasAdiantadas,
	}

	// Calcular média de horas por dia
	if monthStats.DiasTrabalados > 0 {
		mesAtual.MediaHorasDia = monthStats.MinutosTrabalhados / monthStats.DiasTrabalados
	}

	// Calcular taxa de pontualidade
	if monthStats.DiasTrabalados > 0 {
		diasPontuais := monthStats.DiasTrabalados - monthStats.TotalAtrasos
		mesAtual.TaxaPontualidade = (float64(diasPontuais) / float64(monthStats.DiasTrabalados)) * 100
	}

	// Estatísticas gerais
	totalPontos, _ := s.repo.GetTotalPontosCount(userID)
	totalJustificativas, _ := s.repo.GetJustificativasCount(userID, empresaID)
	justificativasPendentes, _ := s.repo.GetJustificativasPendentesCount(userID, empresaID)

	geral := GeralStats{
		TotalPontosRegistrados:  int(totalPontos),
		TotalJustificativas:     int(totalJustificativas),
		JustificativasPendentes: int(justificativasPendentes),
		TaxaPontualizacaoGeral:  mesAtual.TaxaPontualidade, // Pode ser calculado de forma diferente
	}

	// Últimos registros
	recentPontos, _ := s.repo.GetRecentPontos(userID, 10)
	ultimos := make([]RegistroPontoInfo, len(recentPontos))
	for i, ponto := range recentPontos {
		ultimos[i] = RegistroPontoInfo{
			ID:         ponto.ID,
			Data:       ponto.Timestamp,
			Localidade: ponto.Localizacao,
			Tipo:       ponto.Metodo, // "Presencial" ou "Remoto"
		}
	}

	return &StatsResponse{
		BancoHoras:       bancoInfo,
		MesAtual:         mesAtual,
		Geral:            geral,
		UltimosRegistros: ultimos,
	}, nil
}

// GetCalendar retorna o calendário de presença do mês
func (s *service) GetCalendar(userID uint, month, year int) (*CalendarioResponse, error) {
	pontos, err := s.repo.GetPontosByMonth(userID, month, year)
	if err != nil {
		return nil, err
	}

	// Agrupar pontos por dia
	pontoPorDia := make(map[int][]model.RegistroPonto)
	for _, ponto := range pontos {
		dia := ponto.Timestamp.Day()
		pontoPorDia[dia] = append(pontoPorDia[dia], ponto)
	}

	// Criar calendário
	daysInMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()

	// Data de hoje para comparação
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	dias := make([]CalendarioDia, daysInMonth)
	for i := 1; i <= daysInMonth; i++ {
		currentDate := time.Date(year, time.Month(month), i, 0, 0, 0, 0, time.UTC)
		weekday := currentDate.Weekday()

		var status string
		var horasTrabalhadas int
		var tipo string

		if pontosDoDia, existe := pontoPorDia[i]; existe {
			status = "trabalhado"
			// Calcular horas trabalhadas agrupando pares de pontos
			horasTrabalhadas = calculateWorkHoursFromPontos(pontosDoDia)
			tipo = "normal"
		} else if weekday == time.Saturday || weekday == time.Sunday {
			status = "fim_de_semana"
		} else if currentDate.After(today) {
			// Se a data é depois de hoje, é futuro
			status = "futuro"
		} else {
			status = "ausente"
		}

		dias[i-1] = CalendarioDia{
			Dia:              i,
			Status:           status,
			HorasTrabalhadas: horasTrabalhadas,
			Tipo:             tipo,
		}
	}

	return &CalendarioResponse{
		Mes:  month,
		Ano:  year,
		Dias: dias,
	}, nil
}

// GetMyPermissions retorna as permissões detalhadas do usuário
func (s *service) GetMyPermissions(userID uint) (*PermissoesResponse, error) {
	permissoes, err := s.repo.GetPermissoesByUserID(userID)
	if err != nil {
		return nil, err
	}

	detalhadas := make([]PermissaoDetalhada, len(permissoes))
	for i, perm := range permissoes {
		categoria := extractCategory(perm.Nome)
		detalhadas[i] = PermissaoDetalhada{
			ID:          perm.ID,
			Nome:        perm.Nome,
			Descricao:   perm.Descricao,
			Categoria:   categoria,
			PodeVer:     strings.Contains(perm.Nome, "ver") || strings.Contains(perm.Nome, "listar"),
			PodeCriar:   strings.Contains(perm.Nome, "criar") || strings.Contains(perm.Nome, "cadastrar"),
			PodeEditar:  strings.Contains(perm.Nome, "editar") || strings.Contains(perm.Nome, "atualizar"),
			PodeDeletar: strings.Contains(perm.Nome, "deletar") || strings.Contains(perm.Nome, "remover"),
		}
	}

	return &PermissoesResponse{
		Permissoes: detalhadas,
		Total:      len(detalhadas),
	}, nil
}

// GetRecentActivity retorna as atividades recentes do usuário
func (s *service) GetRecentActivity(userID uint, limit int) (*RecentActivityResponse, error) {
	// Buscar últimos pontos
	pontos, _ := s.repo.GetRecentPontos(userID, limit)

	atividades := make([]AtividadeInfo, 0)

	// Converter pontos em atividades
	for _, ponto := range pontos {
		atividades = append(atividades, AtividadeInfo{
			ID:        ponto.ID,
			Tipo:      "ponto",
			Descricao: fmt.Sprintf("Ponto registrado - %s", ponto.Metodo),
			Data:      ponto.Timestamp,
			Icone:     "clock",
			Detalhes: map[string]interface{}{
				"metodo":      ponto.Metodo,
				"localizacao": ponto.Localizacao,
				"latitude":    ponto.Latitude,
				"longitude":   ponto.Longitude,
			},
		})
	}

	return &RecentActivityResponse{
		Atividades: atividades,
		Total:      len(atividades),
	}, nil
}

// Helper functions

func maskCPF(cpf string) string {
	if len(cpf) != 11 {
		return cpf
	}
	return cpf[:3] + ".***.***-" + cpf[9:]
}

func maskCNPJ(cnpj string) string {
	if len(cnpj) != 14 {
		return cnpj
	}
	return cnpj[:2] + ".***.***/" + cnpj[8:12] + "-" + cnpj[12:]
}

func calculateCargaHorariaSemanal(contrato model.Contrato) int {
	if contrato.CargaHorariaSemanalMinutos != nil {
		return int(*contrato.CargaHorariaSemanalMinutos)
	}
	// Padrão: 44h/semana = 2640 minutos
	return 2640
}

func calculateCargaHorariaMensal(contrato model.Contrato) int {
	semanal := calculateCargaHorariaSemanal(contrato)
	// Aproximadamente 4.33 semanas por mês
	return int(float64(semanal) * 4.33)
}

func minutesToTimeString(minutes uint) string {
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%02d:%02d", hours, mins)
}

func formatMinutesToHours(minutes int) string {
	if minutes < 0 {
		minutes = -minutes
		hours := minutes / 60
		mins := minutes % 60
		return fmt.Sprintf("-%02d:%02d", hours, mins)
	}
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("+%02d:%02d", hours, mins)
}

func calculateWorkHoursFromPontos(pontos []model.RegistroPonto) int {
	if len(pontos) < 2 {
		return 0
	}

	// Ordenar por timestamp antes de calcular
	// Criar uma cópia para não modificar o slice original
	pontosCopy := make([]model.RegistroPonto, len(pontos))
	copy(pontosCopy, pontos)

	// Ordenar do mais antigo para o mais recente
	for i := 0; i < len(pontosCopy)-1; i++ {
		for j := i + 1; j < len(pontosCopy); j++ {
			if pontosCopy[i].Timestamp.After(pontosCopy[j].Timestamp) {
				pontosCopy[i], pontosCopy[j] = pontosCopy[j], pontosCopy[i]
			}
		}
	}

	// Calcular horas trabalhadas agrupando pares (entrada/saída)
	totalMinutes := 0
	for i := 0; i < len(pontosCopy)-1; i += 2 {
		entrada := pontosCopy[i].Timestamp
		saida := pontosCopy[i+1].Timestamp
		duracao := saida.Sub(entrada)
		totalMinutes += int(duracao.Minutes())
	}

	return totalMinutes
}

func extractCategory(permissionName string) string {
	lower := strings.ToLower(permissionName)

	categories := map[string]string{
		"ponto":         "ponto",
		"usuario":       "usuario",
		"relatorio":     "relatorio",
		"empresa":       "empresa",
		"cargo":         "cargo",
		"localidade":    "localidade",
		"justificativa": "justificativa",
		"permissao":     "sistema",
		"contrato":      "contrato",
	}

	for key, category := range categories {
		if strings.Contains(lower, key) {
			return category
		}
	}

	return "outros"
}

// UpdateCPF atualiza o CPF de um usuário (apenas admin com permissão EDITAR_USUARIO)
func (s *service) UpdateCPF(adminID, targetUserID uint, req UpdateCPFRequest, ip, userAgent string) error {
	// Validar CPF
	if err := validator.ValidarCPF(req.NovoCPF); err != nil {
		return err
	}

	// Buscar usuário alvo
	targetUser, err := s.repo.GetUserByID(targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("usuário não encontrado")
		}
		return err
	}

	// Guardar CPF antigo para audit log
	cpfAntigo := targetUser.CPF

	// Sanitizar CPF (remover pontos e traços)
	cpfSanitizado := validator.SanitizarCPF(req.NovoCPF)

	// Verificar se CPF já está em uso por outro usuário
	existingUser := &model.Usuario{}
	err = s.db.Where("cpf = ? AND id != ?", cpfSanitizado, targetUserID).First(existingUser).Error
	if err == nil {
		return errors.New("CPF já está em uso por outro usuário")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("erro ao verificar CPF")
	}

	// Atualizar CPF
	targetUser.CPF = cpfSanitizado

	// Salvar alterações
	if err := s.repo.UpdateUser(targetUser); err != nil {
		return errors.New("erro ao atualizar CPF")
	}

	// AUDIT LOG: Registrar alteração de CPF por admin
	var empresaID uint
	if targetUser.Contrato.ID > 0 {
		empresaID = targetUser.Contrato.EmpresaID
	}

	if s.auditLogger != nil {
		dadosAntigos := map[string]interface{}{
			"cpf":            cpfAntigo,
			"admin_id":       adminID,
			"target_user_id": targetUserID,
		}
		dadosNovos := map[string]interface{}{
			"cpf":            cpfSanitizado,
			"admin_id":       adminID,
			"target_user_id": targetUserID,
		}
		_ = s.auditLogger.LogAction(&adminID, &empresaID, "UPDATE_CPF_BY_ADMIN", "usuario", targetUserID, dadosAntigos, dadosNovos, ip, userAgent)
	}

	return nil
}
