// Database type and version selector interaction
// Updates version options when database type changes

window.alpineDatabaseTypeVersion = function () {
  return {
    dbType: "postgresql",
    versions: {
      postgresql: ["18", "17", "16", "15", "14", "13"],
      clickhouse: ["24.3", "24.1", "23.8", "22.8"],
    },
    versionPrefix: {
      postgresql: "PostgreSQL ",
      clickhouse: "ClickHouse ",
    },

    init() {
      // Get initial database type from form if editing
      const dbTypeSelect = this.$el.querySelector(
        'select[name="database_type"]',
      );
      if (dbTypeSelect && dbTypeSelect.value) {
        this.dbType = dbTypeSelect.value;
      }

      // Update version options on init
      this.updateVersionOptions();
    },

    updateDatabaseType() {
      const dbTypeSelect = this.$el.querySelector(
        'select[name="database_type"]',
      );
      if (dbTypeSelect) {
        this.dbType = dbTypeSelect.value;
        this.updateVersionOptions();
      }
    },

    updateVersionOptions() {
      const versionSelect = this.$el.querySelector('select[name="version"]');
      if (!versionSelect) return;

      // Hide/show version field based on database type
      const versionField = versionSelect.closest('.form-control') || versionSelect.parentElement;
      if (versionField) {
        if (this.dbType === "clickhouse") {
          versionField.style.display = "none";
          // Clear version value for ClickHouse
          versionSelect.value = "";
        } else {
          versionField.style.display = "";
        }
      }

      // For ClickHouse, don't update options
      if (this.dbType === "clickhouse") {
        return;
      }

      // Clear existing options except the first placeholder
      const placeholder = versionSelect.querySelector('option[value=""]');
      versionSelect.innerHTML = "";
      if (placeholder) {
        versionSelect.appendChild(placeholder);
      }

      // Add new options based on database type
      const versions = this.versions[this.dbType] || this.versions.postgresql;
      const prefix =
        this.versionPrefix[this.dbType] || this.versionPrefix.postgresql;

      versions.forEach((version) => {
        const option = document.createElement("option");
        option.value = version;
        option.textContent = prefix + version;
        versionSelect.appendChild(option);
      });

      // Reset version selection
      versionSelect.value = "";
    },
  };
};
