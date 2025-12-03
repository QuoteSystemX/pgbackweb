package databases

import (
	"database/sql"
	"reflect"

	"github.com/eduardolat/pgbackweb/internal/database/dbgen"
	"github.com/eduardolat/pgbackweb/internal/util/pathutil"
	"github.com/eduardolat/pgbackweb/internal/validate"
	"github.com/eduardolat/pgbackweb/internal/view/web/component"
	"github.com/eduardolat/pgbackweb/internal/view/web/respondhtmx"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	nodx "github.com/nodxdev/nodxgo"
	alpine "github.com/nodxdev/nodxgo-alpine"
	htmx "github.com/nodxdev/nodxgo-htmx"
	lucide "github.com/nodxdev/nodxgo-lucide"
)

// extractVersionNullString extracts sql.NullString from either string or sql.NullString
// This handles both cases: before SQLC regeneration (string) and after (sql.NullString)
func extractVersionNullString(version interface{}) sql.NullString {
	if version == nil {
		return sql.NullString{Valid: false}
	}

	// Handle sql.NullString
	if ns, ok := version.(sql.NullString); ok {
		return ns
	}

	// Handle string
	if s, ok := version.(string); ok {
		if s == "" {
			return sql.NullString{Valid: false}
		}
		return sql.NullString{String: s, Valid: true}
	}

	// Handle via reflection for other types
	v := reflect.ValueOf(version)
	if v.Kind() == reflect.String {
		s := v.String()
		if s == "" {
			return sql.NullString{Valid: false}
		}
		return sql.NullString{String: s, Valid: true}
	}

	return sql.NullString{Valid: false}
}

// extractVersionString extracts version string from either string or sql.NullString
func extractVersionString(version interface{}) string {
	ns := extractVersionNullString(version)
	if ns.Valid {
		return ns.String
	}
	return ""
}

func (h *handlers) editDatabaseHandler(c echo.Context) error {
	ctx := c.Request().Context()

	databaseID, err := uuid.Parse(c.Param("databaseID"))
	if err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}

	var formData createDatabaseDTO
	if err := c.Bind(&formData); err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}
	if err := validate.Struct(&formData); err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}

	// Version is required only for PostgreSQL
	if formData.DatabaseType == "postgresql" && formData.Version == "" {
		return respondhtmx.ToastError(c, "Version is required for PostgreSQL databases")
	}

	// Prepare version as NullString - NULL for ClickHouse if empty, otherwise the value
	var version sql.NullString
	if formData.DatabaseType == "clickhouse" {
		// For ClickHouse, version can be NULL
		version = sql.NullString{String: "", Valid: false}
	} else {
		// For PostgreSQL, version is required
		version = sql.NullString{String: formData.Version, Valid: true}
	}

	_, err = h.servs.DatabasesService.UpdateDatabase(
		ctx, dbgen.DatabasesServiceUpdateDatabaseParams{
			ID:               databaseID,
			Name:             sql.NullString{String: formData.Name, Valid: true},
			DatabaseType:     sql.NullString{String: formData.DatabaseType, Valid: true},
			Version:          version,
			ConnectionString: sql.NullString{String: formData.ConnectionString, Valid: true},
		},
	)
	if err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}

	return respondhtmx.AlertWithRefresh(c, "Database updated")
}

func editDatabaseButton(
	database dbgen.DatabasesServicePaginateDatabasesRow,
) nodx.Node {
	idPref := "edit-database-" + database.ID.String()
	formID := idPref + "-form"
	btnClass := idPref + "-btn"
	loadingID := idPref + "-loading"

	htmxAttributes := func(url string) nodx.Node {
		return nodx.Group(
			htmx.HxPost(pathutil.BuildPath(url)),
			htmx.HxInclude("#"+formID),
			htmx.HxDisabledELT("."+btnClass),
			htmx.HxIndicator("#"+loadingID),
			htmx.HxValidate("true"),
		)
	}

	mo := component.Modal(component.ModalParams{
		Size:  component.SizeMd,
		Title: "Edit database",
		Content: []nodx.Node{
			nodx.Div(
				alpine.XData("alpineDatabaseTypeVersion()"),
				alpine.XInit("init()"),
				nodx.FormEl(
					nodx.Id(formID),
					nodx.Class("space-y-2"),

				component.InputControl(component.InputControlParams{
					Name:        "name",
					Label:       "Name",
					Placeholder: "My database",
					Required:    true,
					Type:        component.InputTypeText,
					HelpText:    "A name to easily identify the database",
					Children: []nodx.Node{
						nodx.Value(database.Name),
					},
				}),

				component.SelectControl(component.SelectControlParams{
					Name:     "database_type",
					Label:    "Database Type",
					Required: true,
					HelpText: "The type of database",
					Children: []nodx.Node{
						alpine.XModel("dbType"),
						alpine.XOn("change", "updateDatabaseType()"),
						component.DatabaseTypeSelectOptions(sql.NullString{
							Valid:  true,
							String: database.DatabaseType,
						}),
					},
				}),

				component.SelectControl(component.SelectControlParams{
					Name:     "version",
					Label:    "Version",
					Required: false,
					HelpText: "The version of the database",
					Children: []nodx.Node{
						component.DatabaseVersionSelectOptions(database.DatabaseType, extractVersionNullString(database.Version)),
					},
				}),

				component.InputControl(component.InputControlParams{
					Name:        "connection_string",
					Label:       "Connection string",
					Placeholder: "postgresql://user:password@localhost:5432/mydb",
					Required:    true,
					Type:        component.InputTypeText,
					HelpText:    "Connection string for the database. For PostgreSQL: postgresql://user:password@host:port/dbname. For ClickHouse (in Docker): --host=pbw_clickhouse --port=9000 --user=default --password= or clickhouse://default@pbw_clickhouse:9000/default. For local ClickHouse: --host=localhost --port=9000 --user=default --password=. It will be stored securely using PGP encryption.",
					Children: []nodx.Node{
						nodx.Value(database.DecryptedConnectionString),
					},
				}),
				),
			),

			nodx.Div(
				nodx.Class("flex justify-between items-center pt-4"),
				nodx.Div(
					nodx.Button(
						htmxAttributes("/dashboard/databases/test"),
						nodx.ClassMap{
							btnClass:                      true,
							"btn btn-neutral btn-outline": true,
						},
						nodx.Type("button"),
						component.SpanText("Test connection"),
						lucide.DatabaseZap(),
					),
				),
				nodx.Div(
					nodx.Class("flex justify-end items-center space-x-2"),
					component.HxLoadingMd(loadingID),
					nodx.Button(
						htmxAttributes("/dashboard/databases/"+database.ID.String()+"/edit"),
						nodx.ClassMap{
							btnClass:          true,
							"btn btn-primary": true,
						},
						nodx.Type("button"),
						component.SpanText("Save"),
						lucide.Save(),
					),
				),
			),
		},
	})

	return nodx.Div(
		mo.HTML,
		component.OptionsDropdownButton(
			mo.OpenerAttr,
			lucide.Pencil(),
			component.SpanText("Edit database"),
		),
	)
}
