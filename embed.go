package lynway_wms

import "embed"

//go:embed migrations/wms_db/*.sql
var MigrationsFS embed.FS
