package email

const (
	TemplatePasswordReset = "password_reset"
	TemplateWelcome       = "welcome"
)

func GetTemplate(name string) string {
	switch name {
	case TemplatePasswordReset:
		return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Recuperação de Senha</title>
    <style>
        body { margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #0f172a; color: #e2e8f0; }
        .container { max-width: 600px; margin: 40px auto; background: #1e293b; border-radius: 16px; overflow: hidden; box-shadow: 0 10px 25px rgba(0,0,0,0.3); border: 1px solid #334155; }
        .header { background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%); padding: 30px; text-align: center; border-bottom: 1px solid #334155; }
        .logo { font-size: 28px; font-weight: bold; color: #ffffff; text-decoration: none; display: inline-block; }
        .logo span { color: #fbbf24; } /* Nexora Gold */
        .content { padding: 40px 30px; text-align: center; }
        h1 { color: #ffffff; font-size: 24px; margin-bottom: 20px; font-weight: 600; }
        p { color: #94a3b8; font-size: 16px; line-height: 1.6; margin-bottom: 24px; }
        .button { display: inline-block; background: linear-gradient(to right, #fbbf24, #d97706); color: #000000; padding: 14px 32px; text-decoration: none; border-radius: 8px; font-weight: bold; font-size: 16px; transition: transform 0.2s; box-shadow: 0 4px 6px rgba(251, 191, 36, 0.2); }
        .button:hover { transform: translateY(-2px); box-shadow: 0 6px 8px rgba(251, 191, 36, 0.3); }
        .link-box { margin-top: 30px; padding: 15px; background: #0f172a; border-radius: 8px; border: 1px solid #334155; word-break: break-all; }
        .link-text { color: #38bdf8; font-size: 13px; text-decoration: none; }
        .footer { background-color: #0f172a; padding: 20px; text-align: center; font-size: 12px; color: #64748b; border-top: 1px solid #334155; }
        .warning { font-size: 13px; color: #ef4444; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">Nexora<span>Ponto</span></div>
        </div>
        <div class="content">
            <h1>Redefinição de Senha</h1>
            <p>Olá, <strong>{{.Name}}</strong>!</p>
            <p>Recebemos uma solicitação para redefinir a senha da sua conta. Para continuar e criar uma nova senha, clique no botão abaixo:</p>
            
            <a href="{{.Link}}" class="button">Redefinir Minha Senha</a>

            <p class="warning">Se você não solicitou esta alteração, por favor ignore este email. Sua senha permanecerá a mesma.</p>
            
            <div class="link-box">
                <p style="margin: 0 0 10px 0; font-size: 12px; color: #64748b;">Ou copie e cole o link abaixo no seu navegador:</p>
                <a href="{{.Link}}" class="link-text">{{.Link}}</a>
            </div>
            
            <p style="font-size: 12px; margin-top: 20px; color: #64748b;">Este link é válido por 1 hora.</p>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} Nexora Ponto. Todos os direitos reservados.</p>
            <p>Este é um email automático, por favor não responda.</p>
        </div>
    </div>
</body>
</html>`
	case TemplateWelcome:
		return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Bem-vindo ao Nexora</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f4f4; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 20px auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        .header { background-color: #1a1a1a; padding: 20px; text-align: center; }
        .header h1 { color: #ffe27a; margin: 0; font-size: 24px; }
        .content { padding: 30px; color: #333333; line-height: 1.6; }
        .button { display: inline-block; background-color: #ffe27a; color: #000000; padding: 12px 24px; text-decoration: none; border-radius: 4px; font-weight: bold; margin-top: 20px; }
        .footer { background-color: #f9f9f9; padding: 15px; text-align: center; font-size: 12px; color: #888888; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Bem-vindo ao Nexora</h1>
        </div>
        <div class="content">
            <h2>Olá, {{.Name}}!</h2>
            <p>Sua conta foi criada com sucesso. Estamos muito felizes em tê-lo conosco.</p>
            <p>Acesse sua conta agora mesmo para começar:</p>
            <center>
                <a href="{{.Link}}" class="button">Acessar Sistema</a>
            </center>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} Nexora Ponto. Todos os direitos reservados.</p>
        </div>
    </div>
</body>
</html>`
	default:
		return ""
	}
}
