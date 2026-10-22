package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/dustin/go-humanize"
	"github.com/urfave/cli/v3"
	_ "modernc.org/sqlite"
)

func doConvert(ctx context.Context, cmd *cli.Command) error {
	inputFile := cmd.StringArg(inputFileArg.Name)
	outputFile := cmd.StringArg(outputFileArg.Name)

	if inputFile == "" {
		return fmt.Errorf("%s is required", inputFileArg.Name)
	}
	if outputFile == "" {
		outputFile = inputFile + ".csv"
	}

	db, err := sql.Open("sqlite", inputFile)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	mimeMap := make(map[int64]string)
	err = func() error {
		query := sq.Select("id", "name").
			From("mime")

		rows, err := query.RunWith(db).QueryContext(ctx)
		if err != nil {
			return err
		}
		defer func(rows *sql.Rows) {
			_ = rows.Close()
		}(rows)

		for rows.Next() {
			var id int64
			var name string

			err = rows.Scan(&id, &name)
			if err != nil {
				return err
			}
			mimeMap[id] = name
		}
		if rows.Err() != nil {
			return rows.Err()
		}

		return nil
	}()

	query := sq.Select("id", "parent", "mime", "path", "version", "mtime", "size", "thumbnail_count", "json_data").
		From("document").
		OrderBy("path ASC")

	rows, err := query.RunWith(db).QueryContext(ctx)
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	csvWriter := csv.NewWriter(f)
	err = csvWriter.Write([]string{
		"id", "parent", "path", "mtime_formatted", "size_formatted", "mtime", "size", "mime",
	})
	if err != nil {
		return err
	}
	defer csvWriter.Flush()

	for rows.Next() {
		var id int64
		var parentId *int64
		var mimeId *int64
		var path string
		var version int64
		var mtime int64
		var size int64
		var thumbnailCount int64
		var jsonData *string

		err = rows.Scan(&id, &parentId, &mimeId, &path, &version, &mtime, &size, &thumbnailCount, &jsonData)
		if err != nil {
			return err
		}

		var mime string
		if mimeId != nil {
			mime = mimeMap[*mimeId]
		}

		err = csvWriter.Write([]string{
			formatInt64(id),
			formatInt64Ptr(parentId),
			path,
			formatTime(mtime),
			formatSize(size),
			formatInt64(mtime),
			formatInt64(size),
			mime,
		})
		if err != nil {
			return err
		}
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	return nil
}

func formatInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}

func formatInt64Ptr(v *int64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatInt(*v, 10)
}

func formatTime(v int64) string {
	t := time.Unix(v, 0).Local()
	return t.Format(time.RFC3339)
}

func formatSize(v int64) string {
	return humanize.Bytes(uint64(v))
}
