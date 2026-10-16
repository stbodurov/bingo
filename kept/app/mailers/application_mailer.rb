class ApplicationMailer < ActionMailer::Base
  default from: ENV.fetch("MAIL_FROM", "Kept <hello@kept.example>")
  layout "mailer"
end
