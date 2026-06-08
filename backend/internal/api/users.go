type User struct {
    ID        string    `json:"id"`
    UUID      string    `json:"uuid"`
    Email     string    `json:"email"`
    TrafficGB int64     `json:"traffic_gb"`
    UsedGB    float64   `json:"used_gb"`
    ExpiresAt time.Time `json:"expires_at"`
    Active    bool      `json:"active"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    json.NewDecoder(r.Body).Decode(&req)
    user := &models.User{
        UUID:      uuid.New().String(),
        Email:     req.Email,
        TrafficGB: req.TrafficGB,
        ExpiresAt: time.Now().AddDate(0, 0, req.ExpireDays),
    }
    h.db.Create(user)
    h.core.Reload()  // перегенерировать конфиг ядра
    json.NewEncoder(w).Encode(user)
}