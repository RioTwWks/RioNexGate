type CoreManager interface {
    Reload() error
    GetStats(userID string) (float64, error)
    GetClientLink(userID string, protocol string) string
}

type XrayManager struct {
    configPath string
    apiClient  *xrayAPI.Client
}

func (x *XrayManager) Reload() error {
    // генерация полного config.json из шаблона + список пользователей из БД
    // затем вызов `kill -SIGHUP` или API xray
}

type SingBoxManager struct { ... } // аналогично