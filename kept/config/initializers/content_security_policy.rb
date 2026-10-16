Rails.application.configure do
  config.content_security_policy do |policy|
    policy.default_src :self
    policy.font_src    :self, "https://fonts.gstatic.com"
    policy.img_src     :self, :data
    policy.object_src  :none
    policy.script_src  :self
    policy.style_src   :self, "https://fonts.googleapis.com"
    policy.style_src_attr :unsafe_inline # inline style attributes on SVG rings/bars only; no inline <style> without a nonce
    policy.connect_src :self
    policy.frame_ancestors :none
    policy.base_uri :self
    policy.form_action :self
  end
  # Nonces for the importmap and any inline <script>/<style> Rails generates.
  config.content_security_policy_nonce_generator = ->(_request) { SecureRandom.base64(16) }
  config.content_security_policy_nonce_directives = %w[script-src style-src]
  config.content_security_policy_nonce_auto = true
end
