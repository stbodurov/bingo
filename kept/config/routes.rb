Rails.application.routes.draw do
  root "pages#home"
  get "employers", to: "pages#employers"
  get "science", to: "pages#science"
  get "privacy", to: "pages#privacy"

  post "waitlist", to: "waitlist_signups#create", as: :waitlist_signups
  get "waitlist/:code", to: "waitlist_signups#show", as: :waitlist_status

  resource :session
  resources :passwords, param: :token
  resource :registration, only: %i[new create]
  post "demo", to: "sessions#demo", as: :demo_session

  scope "app" do
    get "/", to: "today#show", as: :today
    resource :medication, only: %i[new create edit update]
    resources :doses, only: %i[index create destroy]
    resources :symptoms, only: %i[index create destroy]
    get "nutrition", to: "nutrition#show", as: :nutrition
    resources :foods, only: %i[create destroy]
    resources :waters, only: :create
    get "strength", to: "strength#show", as: :strength
    resources :workouts, only: %i[create destroy]
    get "progress", to: "progress#show", as: :progress
    resources :weights, only: %i[create destroy]
    resources :check_ins, only: :create
    get "plan", to: "plan#show", as: :plan
    resource :off_ramp, only: %i[create destroy] do
      post :step
      post :stop
    end
    get "report", to: "reports#show", as: :report
    get "appeal", to: "reports#appeal", as: :appeal
    resource :profile, only: %i[edit update destroy]
    post "profile/enroll", to: "profiles#enroll", as: :enroll_profile
    delete "profile/enroll", to: "profiles#unenroll"
    get "export", to: "exports#show", as: :export
    post "calendar/reset", to: "calendars#reset", as: :reset_calendar
  end

  get "calendar/:token", to: "calendars#show", as: :dose_calendar, constraints: { format: :ics }
  get "employer/:slug", to: "employer_dashboards#show", as: :employer_dashboard
  post "employer/:slug/access", to: "employer_dashboards#access", as: :employer_access

  get "up" => "rails/health#show", as: :rails_health_check
  get "manifest" => "rails/pwa#manifest", as: :pwa_manifest
  get "service-worker" => "rails/pwa#service_worker", as: :pwa_service_worker
end
