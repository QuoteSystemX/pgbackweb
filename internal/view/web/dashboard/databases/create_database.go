package databases

import (
	"database/sql"

	"github.com/eduardolat/pgbackweb/internal/database/dbgen"
	"github.com/eduardolat/pgbackweb/internal/util/pathutil"
	"github.com/eduardolat/pgbackweb/internal/validate"
	"github.com/eduardolat/pgbackweb/internal/view/web/component"
	"github.com/eduardolat/pgbackweb/internal/view/web/respondhtmx"
	"github.com/labstack/echo/v4"
	nodx "github.com/nodxdev/nodxgo"
	alpine "github.com/nodxdev/nodxgo-alpine"
	htmx "github.com/nodxdev/nodxgo-htmx"
	lucide "github.com/nodxdev/nodxgo-lucide"
)

type createDatabaseDTO struct {
	Name             string `form:"name" validate:"required"`
	DatabaseType     string `form:"database_type" validate:"required"`
	Version          string `form:"version"`
	ConnectionString string `form:"connection_string" validate:"required"`
}

func (h *handlers) createDatabaseHandler(c echo.Context) error {
	ctx := c.Request().Context()

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

	// For ClickHouse, use empty string (will be stored as NULL in DB)
	version := formData.Version
	if formData.DatabaseType == "clickhouse" {
		version = ""
	}

	_, err := h.servs.DatabasesService.CreateDatabase(
		ctx, dbgen.DatabasesServiceCreateDatabaseParams{
			Name:             formData.Name,
			DatabaseType:     formData.DatabaseType,
			Version:          version,
			ConnectionString: formData.ConnectionString,
		},
	)
	if err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}

	return respondhtmx.Redirect(c, pathutil.BuildPath("/dashboard/databases"))
}

func createDatabaseButton() nodx.Node {
	htmxAttributes := func(url string) nodx.Node {
		return nodx.Group(
			htmx.HxPost(pathutil.BuildPath(url)),
			htmx.HxInclude("#add-database-form"),
			htmx.HxDisabledELT(".add-database-btn"),
			htmx.HxIndicator("#add-database-loading"),
			htmx.HxValidate("true"),
		)
	}

	mo := component.Modal(component.ModalParams{
		Size:  component.SizeMd,
		Title: "Add database",
			Content: []nodx.Node{
			nodx.Div(
				alpine.XData("alpineDatabaseTypeVersion()"),
				alpine.XInit("init()"),
				nodx.FormEl(
					nodx.Id("add-database-form"),
					nodx.Class("space-y-2"),

				component.InputControl(component.InputControlParams{
					Name:        "name",
					Label:       "Name",
					Placeholder: "My database",
					Required:    true,
					Type:        component.InputTypeText,
					HelpText:    "A name to easily identify the database",
				}),

				component.SelectControl(component.SelectControlParams{
					Name:        "database_type",
					Label:       "Database Type",
					Placeholder: "Select a database type",
					Required:    true,
					HelpText:    "The type of database",
					Children: []nodx.Node{
						alpine.XModel("dbType"),
						alpine.XOn("change", "updateDatabaseType()"),
						component.DatabaseTypeSelectOptions(sql.NullString{}),
					},
				}),

				component.SelectControl(component.SelectControlParams{
					Name:        "version",
					Label:       "Version",
					Placeholder: "Select a version",
					Required:    false,
					HelpText:    "The version of the database",
					Children: []nodx.Node{
						component.DatabaseVersionSelectOptions("postgresql", sql.NullString{}),
					},
				}),

				component.InputControl(component.InputControlParams{
					Name:        "connection_string",
					Label:       "Connection string",
					Placeholder: "postgresql://user:password@localhost:5432/mydb",
					Required:    true,
					Type:        component.InputTypeText,
					HelpText:    "Connection string for the database. For PostgreSQL: postgresql://user:password@host:port/dbname. For ClickHouse (in Docker): --host=pbw_clickhouse --port=9000 --user=default --password= or clickhouse://default@pbw_clickhouse:9000/default. For local ClickHouse: --host=localhost --port=9000 --user=default --password=. It will be stored securely using PGP encryption.",
				}),
				),
			),

			nodx.Div(
				nodx.Class("flex justify-between items-center pt-4"),
				nodx.Div(
					nodx.Button(
						htmxAttributes("/dashboard/databases/test"),
						nodx.Class("add-database-btn btn btn-neutral btn-outline"),
						nodx.Type("button"),
						component.SpanText("Test connection"),
						lucide.DatabaseZap(),
					),
				),
				nodx.Div(
					nodx.Class("flex justify-end items-center space-x-2"),
					component.HxLoadingMd("add-database-loading"),
					nodx.Button(
						htmxAttributes("/dashboard/databases"),
						nodx.Class("add-database-btn btn btn-primary"),
						nodx.Type("button"),
						component.SpanText("Add database"),
						lucide.Save(),
					),
				),
			),
		},
	})

	button := nodx.Button(
		mo.OpenerAttr,
		nodx.Class("btn btn-primary"),
		component.SpanText("Add database"),
		lucide.Plus(),
	)

	return nodx.Div(
		nodx.Class("inline-block"),
		mo.HTML,
		button,
	)
}
