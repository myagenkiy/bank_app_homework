package services

import (
	"bank-api/internal/models"
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/go-mail/mail/v2"
)

type EmailService struct {
	dialer *mail.Dialer
	from   string
}

func NewEmailService() *EmailService {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 587
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if smtpHost == "" {
		smtpHost = "smtp.gmail.com" // можно заменить на другой SMTP
	}
	if from == "" {
		from = "bank@api.com"
	}

	dialer := mail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	dialer.TLSConfig = &tls.Config{
		InsecureSkipVerify: false,
	}

	return &EmailService{
		dialer: dialer,
		from:   from,
	}
}

// sendEmail отправляет письмо
func (s *EmailService) sendEmail(to, subject, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if err := s.dialer.DialAndSend(m); err != nil {
		log.Printf("SMTP error: %v", err)
		return fmt.Errorf("email sending failed: %v", err)
	}

	log.Printf("Email sent to %s", to)
	return nil
}

// SendPaymentNotification отправляет уведомление о платеже
func (s *EmailService) SendPaymentNotification(notification *models.PaymentNotification) error {
	var subject, body string

	switch notification.Type {
	case "deposit":
		subject = "💰 Пополнение счета"
		body = fmt.Sprintf(`
			<h1>Здравствуйте, %s!</h1>
			<p>Ваш счет был пополнен на сумму <strong>%.2f RUB</strong>.</p>
			<p>Статус: <strong style="color:green;">%s</strong></p>
			<p>Описание: %s</p>
			<hr>
			<small>Это автоматическое уведомление от банка.</small>
		`, notification.UserName, notification.Amount, notification.Status, notification.Description)

	case "transfer":
		subject = "💸 Перевод средств"
		body = fmt.Sprintf(`
			<h1>Здравствуйте, %s!</h1>
			<p>Был совершен перевод на сумму <strong>%.2f RUB</strong>.</p>
			<p>Статус: <strong style="color:green;">%s</strong></p>
			<p>Описание: %s</p>
			<hr>
			<small>Это автоматическое уведомление от банка.</small>
		`, notification.UserName, notification.Amount, notification.Status, notification.Description)

	case "credit_payment":
		subject = "🏦 Платеж по кредиту"
		body = fmt.Sprintf(`
			<h1>Здравствуйте, %s!</h1>
			<p>Произведен платеж по кредиту на сумму <strong>%.2f RUB</strong>.</p>
			<p>Статус: <strong style="color:green;">%s</strong></p>
			<hr>
			<small>Это автоматическое уведомление от банка.</small>
		`, notification.UserName, notification.Amount, notification.Status)

	case "overdue":
		subject = "⚠️ Просрочка платежа по кредиту"
		body = fmt.Sprintf(`
			<h1>Уважаемый(ая) %s!</h1>
			<p style="color:red;">У вас образовалась просрочка по кредиту!</p>
			<p>Сумма просроченного платежа: <strong>%.2f RUB</strong></p>
			<p>Пожалуйста, пополните счет для автоматического списания.</p>
			<hr>
			<small>Это автоматическое уведомление от банка.</small>
		`, notification.UserName, notification.Amount)
	}

	return s.sendEmail(notification.UserEmail, subject, body)
}

// SendWelcomeEmail отправляет приветственное письмо
func (s *EmailService) SendWelcomeEmail(email, username string) error {
	subject := "Добро пожаловать в Банк!"
	body := fmt.Sprintf(`
		<h1>Здравствуйте, %s!</h1>
		<p>Рады приветствовать вас в нашем банке!</p>
		<p>Вы успешно зарегистрировались в системе.</p>
		<p>Теперь вам доступны следующие возможности:</p>
		<ul>
			<li>Открытие счетов в разных валютах</li>
			<li>Выпуск банковских карт</li>
			<li>Переводы между счетами</li>
			<li>Оформление кредитов</li>
		</ul>
		<hr>
		<small>Это автоматическое уведомление от банка.</small>
	`, username)

	return s.sendEmail(email, subject, body)
}
