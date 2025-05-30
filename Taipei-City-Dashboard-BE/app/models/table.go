package models

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"regexp"
	"strconv"
	"strings"
)

// ColumnDef defines dynamic table column with extra options.
type ColumnDef struct {
    Name         string `json:"name"`
    Type         string `json:"type"`
    NotNull      bool   `json:"not_null"`
    IsPrimaryKey bool   `json:"is_primary_key"`
    Comment      string `json:"comment"`
}

// SQL keyword list (should be expanded in production)
var allowedTypes = map[string]bool{
    "serial": true,
    "bigserial": true,
    "smallserial": true,
    "int": true,
    "integer": true,
    "bigint": true,
    "smallint": true,
    "float": true,
    "float4": true,
    "float8": true,
    "double": true,
    "double precision": true,
    "real": true,
    "numeric": true,
    "decimal": true,
    "money": true,
    "varchar": true,
    "character varying": true,
    "char": true,
    "character": true,
    "text": true,
    "boolean": true,
    "bool": true,
    "date": true,
    "timestamp": true,
    "timestamp without time zone": true,
    "timestamp with time zone": true,
    "time": true,
    "time without time zone": true,
    "time with time zone": true,
    "json": true,
    "jsonb": true,
}

var sqlKeywords = map[string]bool{
    "select": true,
    "from": true,
    "where": true,
    "insert": true,
    "update": true,
    "delete": true,
    "create": true,
    "alter": true,
    "drop": true,
    "table": true,
    "index": true,
    "view": true,
    "join": true,
    "inner": true,
    "left": true,
    "right": true,
    "full": true,
    "on": true,
    "as": true,
    "and": true,
    "or": true,
    "not": true,
    "null": true,
    "is": true,
    "into": true,
    "values": true,
    "set": true,
    "primary": true,
    "key": true,
    "foreign": true,
    "constraint": true,
    "unique": true,
    "check": true,
    "default": true,
    "order": true,
    "by": true,
    "group": true,
    "having": true,
    "union": true,
    "all": true,
    "exists": true,
    "case": true,
    "when": true,
    "then": true,
    "else": true,
    "end": true,
    "distinct": true,
    "limit": true,
    "offset": true,
    "true": true,
    "false": true,
    "user": true,
    "grant": true,
    "revoke": true,
    "cast": true,
    "column": true,
    "if": true,
    "do": true,
    "between": true,
    "in": true,
    "like": true,
    "database": true,
    "sequence": true,
    "int": true,
    "integer": true,
    "smallint": true,
    "bigint": true,
    "float": true,
    "double": true,
    "decimal": true,
    "numeric": true,
    "real": true,
    "serial": true,
    "bigserial": true,
    "boolean": true,
    "date": true,
    "time": true,
    "timestamp": true,
    "interval": true,
    "text": true,
    "varchar": true,
    "char": true,
    "json": true,
    "jsonb": true,
}

var typeMustHaveParam = map[string]bool{
    "varchar": true,
    "character varying": true,
    "char": true,
    "character": true,
}

/*
ImportTableFromCSV performs all validation, table creation, and data import logic.
Any failure will return error for controller to handle.
*/
func ImportTableFromCSV(tableName string, columnsJson string, file multipart.File) error {
    // 1. Validate tableName
	if err := validateTableName(tableName); err != nil {
        return err
    }

    // 2. Parse and validate columns
    columns, err := parseAndValidateColumns(columnsJson)
    if err != nil {
        return err
    }

    // 3. Create table (with comments and pk)
    fileBytes, err := io.ReadAll(file)
    if err != nil {
        return fmt.Errorf("failed to read uploaded CSV: %v", err)
    }
    if err := readAndCheckCSVHeader(fileBytes, columns); err != nil {
        return err
    }
	// TODO: skipped
    if err := checkCSVDataRows(fileBytes, columns, 5); err != nil {
        return err
    }

    columns, pkAdded := ensurePK(columns)

    if err := createTableSQL(tableName, columns); err != nil {
        return err
    }
    if err := importCSVRows(tableName, columns, fileBytes, pkAdded); err != nil {
        return err
    }
    return nil
}

func validateTableName(name string) error {
    if name == "" {
        return fmt.Errorf("tableName is required")
    }
    if !isValidTableName(name) {
        return fmt.Errorf("invalid table name")
    }
    if isSQLKeyword(name) {
        return fmt.Errorf("table name is reserved keyword")
    }
    if isTableExists(name) {
        return fmt.Errorf("table already exists")
    }
    return nil
}

func parseAndValidateColumns(columnsJson string) ([]ColumnDef, error) {
    var columns []ColumnDef
    if err := json.Unmarshal([]byte(columnsJson), &columns); err != nil {
        return nil, fmt.Errorf("columns JSON parse error: %v", err)
    }
    if len(columns) == 0 {
        return nil, fmt.Errorf("at least one column is required")
    }

	typeWithParamRegexp := regexp.MustCompile(`^([a-zA-Z ]+)\s*\(([^)]+)\)$`)

	extractBaseTypeAndParam := func(typeStr string) (baseType, param string) {
        typeStr = strings.ToLower(strings.TrimSpace(typeStr))
        match := typeWithParamRegexp.FindStringSubmatch(typeStr)
        if len(match) == 3 {
            return strings.TrimSpace(match[1]), strings.TrimSpace(match[2])
        }
        idx := strings.Index(typeStr, "(")
        if idx != -1 {
            return strings.TrimSpace(typeStr[:idx]), ""
        }
        return typeStr, ""
    }

    names := map[string]bool{}
    for i, col := range columns {
        if col.Name == "" || !isValidTableName(col.Name) {
            return nil, fmt.Errorf("invalid column name: %s", col.Name)
        }
        if isSQLKeyword(col.Name) {
            return nil, fmt.Errorf("column name '%s' is reserved keyword", col.Name)
        }
        if names[col.Name] {
            return nil, fmt.Errorf("duplicate column name: %s", col.Name)
        }
        names[col.Name] = true
        if col.Type == "" {
            return nil, fmt.Errorf("column '%s' type required", col.Name)
        }

		baseType, param := extractBaseTypeAndParam(col.Type)
        if !allowedTypes[baseType] {
            return nil, fmt.Errorf("unsupported column type: %s", col.Type)
        }
        if typeMustHaveParam[baseType] && param == "" {
            return nil, fmt.Errorf("column '%s' of type '%s' must specify length, e.g. '%s(255)'", col.Name, baseType, baseType)
        }
        if typeMustHaveParam[baseType] && param != "" {
            n, err := strconv.Atoi(param)
            if err != nil || n <= 0 {
                return nil, fmt.Errorf("column '%s' of type '%s' must have positive integer length, got '%s'", col.Name, baseType, param)
            }
            if n > 4096 {
                return nil, fmt.Errorf("column '%s' of type '%s' length too long: %d", col.Name, baseType, n)
            }
        }

        if !isValidComment(col.Comment) {
            return nil, fmt.Errorf("invalid comment for column '%s'", col.Name)
        }
        columns[i].Type = strings.ToUpper(col.Type)
    }

    return columns, nil
}

func readAndCheckCSVHeader(fileBytes []byte, columns []ColumnDef) error {
    reader := csv.NewReader(strings.NewReader(string(fileBytes)))
    header, err := reader.Read()
    if err != nil {
        return fmt.Errorf("CSV header read failed: %v", err)
    }

    expectCols := []string{}
    for _, c := range columns {
		expectCols = append(expectCols, c.Name)
    }
    if len(header) != len(expectCols) {
        return fmt.Errorf("CSV header column count mismatch: expect %d, got %d", len(expectCols), len(header))
    }
    for i, h := range header {
        if h != expectCols[i] {
            return fmt.Errorf("CSV header mismatch at column %d: expect '%s', got '%s'", i+1, expectCols[i], h)
        }
    }
    return nil
}

// TODO: skipped
func checkCSVDataRows(fileBytes []byte, columns []ColumnDef, n int) error {
    return nil
}

func isSQLKeyword(word string) bool {
    return sqlKeywords[strings.ToLower(word)]
}

// isValidTableName checks if the name consists only of English letters, numbers, or underscores ([A-Za-z0-9_])
func isValidTableName(name string) bool {
    if name == "" { return false }
    for _, c := range name {
        if !(c == '_' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
            return false
        }
    }
    return true
}

// isTableExists returns true if table already exists.
func isTableExists(table string) bool {
    var count int64
    DBDashboard.Raw("SELECT COUNT(*) FROM pg_tables WHERE schemaname='public' AND tablename=?", table).Scan(&count)
    return count > 0
}

// ensurePK auto-adds id SERIAL PRIMARY KEY if none exists.
func ensurePK(columns []ColumnDef) ([]ColumnDef, bool) {
    hasPK := false
    for _, c := range columns {
        if c.IsPrimaryKey {
            hasPK = true
            break
        }
    }
    if !hasPK {
        return append([]ColumnDef{{Name: "id", Type: "SERIAL", NotNull: true, IsPrimaryKey: true, Comment: "Auto ID"}}, columns...), true
    }
    return columns, false
}

func createTableSQL(tableName string, columns []ColumnDef) error {
    sql := buildCreateTableSQL(tableName, columns)
    if err := DBDashboard.Exec(sql).Error; err != nil {
        return fmt.Errorf("create table failed: %v", err)
    }
    return nil
}

// buildCreateTableSQL generates CREATE TABLE, PK, and COMMENT SQL.
func buildCreateTableSQL(table string, columns []ColumnDef) string {
    var fields []string
    var pk []string
    var comments []string

    for _, c := range columns {
        field := fmt.Sprintf(`"%s" %s`, c.Name, c.Type)
        if c.NotNull {
            field += " NOT NULL"
        }
        fields = append(fields, field)
        if c.IsPrimaryKey {
            pk = append(pk, c.Name)
        }
        if c.Comment != "" {
            comments = append(comments, fmt.Sprintf(
                `COMMENT ON COLUMN "%s"."%s" IS '%s';`, table, c.Name, escapeSQLString(c.Comment)))
        }
    }

    create := fmt.Sprintf(`CREATE TABLE "%s" (%s`, table, strings.Join(fields, ", "))
    if len(pk) > 0 {
        create += fmt.Sprintf(", PRIMARY KEY (%s)", quoteJoin(pk))
    }
    create += ");"
    if len(comments) > 0 {
        create += "\n" + strings.Join(comments, "\n")
    }

    return create
}

// quoteJoin safely joins identifiers with double quotes.
// Notice: use this function only after verifying each name with isValidTableName()
func quoteJoin(cols []string) string {
    for i, c := range cols { cols[i] = `"` + c + `"` }
    return strings.Join(cols, ",")
}

/*
isValidComment validates a SQL column/table comment for safety and cleanliness.

1. Limit maximum length to 255 characters (most databases restrict comment length; too long causes errors or wastes space).
2. Forbid newline (\n, \r): prevents breaking SQL syntax or injection via line breaks.
3. Forbid NULL (0x00): never valid in SQL strings; DBMS treats it as a terminator.
4. Forbid semicolon (;): prevents SQL injection and multi-statement attacks.
5. Forbid double quote (") and backtick (`): avoids interfering with SQL identifier quoting or accidental syntax breakage.
6. Forbid ASCII control characters (0x01~0x1F, except space and optionally tab): disallow invisible/control codes.
7. Allow regular English/Chinese/alphanumeric/punctuation; can extend further as needed.
*/
func isValidComment(comment string) bool {
    if len(comment) == 0 {
        // Empty comments are allowed; set to false if not desired.
        return true
    }
    if len(comment) > 255 {
        return false
    }
    for _, c := range comment {
        // Forbid NULL, newlines, semicolon, double quote, backtick
        switch c {
        case '\x00', '\n', '\r', ';', '"', '`':
            return false
        }
        // Forbid ASCII control characters (0x01~0x1F), except optionally allowing tab (\t, 0x09)
        // if c < 0x20 && c != 0x09 {
        if c < 0x20 {
            return false
        }
    }
    return true
}

// escapeSQLString prevents single quote SQL injection in comments.
func escapeSQLString(str string) string {
    return strings.ReplaceAll(str, "'", "''")
}

// importCSVRows reads and inserts all CSV rows, checking headers match columns.
func importCSVRows(tableName string, columns []ColumnDef, fileBytes []byte, pkAdded bool) error {
    reader := csv.NewReader(strings.NewReader(string(fileBytes)))
	_, err := reader.Read() // drop header
    if err != nil {
        return fmt.Errorf("CSV header read failed: %v", err)
    }

	insertCols := []string{}
    if pkAdded {
        for _, c := range columns {
            if !(strings.ToLower(c.Name) == "id" && c.IsPrimaryKey) {
                insertCols = append(insertCols, c.Name)
            }
        }
    } else {
        for _, c := range columns {
            insertCols = append(insertCols, c.Name)
        }
    }

    placeholders := make([]string, len(insertCols))
    for i := range insertCols { placeholders[i] = "?" }

    tx := DBDashboard.Begin()
    for {
        record, err := reader.Read()
        if err == io.EOF { break }
        if err != nil { tx.Rollback(); return err }
        values := toInterfaceSlice(record)
        sql := fmt.Sprintf(
			`INSERT INTO "%s" (%s) VALUES (%s)`, 
			tableName,
            `"` + strings.Join(insertCols, `","`) + `"`, 
			strings.Join(placeholders, ","),
		)
        if err := tx.Exec(sql, values...).Error; err != nil {
            tx.Rollback()
            return fmt.Errorf("insert row failed: %v", err)
        }
    }
    return tx.Commit().Error
}

// toInterfaceSlice: string[] → interface[]{}
func toInterfaceSlice(strs []string) []interface{} {
    out := make([]interface{}, len(strs))
    for i, s := range strs { out[i] = s }
    return out
}

// ListAllTableNames returns all table names in public schema.
func ListAllTableNames() ([]string, error) {
    var names []string
    rows, err := DBDashboard.Raw(`SELECT tablename FROM pg_tables WHERE schemaname='public'`).Rows()
    if err != nil { return nil, err }
    defer rows.Close()
    for rows.Next() {
        var t string
        rows.Scan(&t)
        names = append(names, t)
    }
    return names, nil
}

// GetSampleRows returns sample rows and columns for a table.
func GetSampleRows(table string, limit int) ([]string, []map[string]interface{}, error) {
    rows, err := DBDashboard.Raw(fmt.Sprintf(`SELECT * FROM "%s" LIMIT %d`, table, limit)).Rows()
    if err != nil { return nil, nil, err }
    defer rows.Close()
    cols, _ := rows.Columns()
    results := []map[string]interface{}{}
    for rows.Next() {
        values := make([]interface{}, len(cols))
        refs := make([]interface{}, len(cols))
        for i := range values { refs[i] = &values[i] }
        rows.Scan(refs...)
        row := map[string]interface{}{}
        for i, col := range cols { row[col] = values[i] }
        results = append(results, row)
    }
    return cols, results, nil
}