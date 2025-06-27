package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Platinaa777/DataCat/internal/entity"
)

var searchFields = map[string]map[string]bool{
	"dashboard":   {"id": true, "name": true, "description": true, "tags": true, "createdAt": true, "updatedAt": true},
	"alert":       {"id": true, "description": true, "status": true, "tags": true, "rawQuery": true, "dataSourceId": true, "previousExecutionTime": true, "nextExecutionTime": true},
	"data-source": {"id": true, "name": true, "type": true, "typeId": true, "purpose": true},
	"plugin":      {"pluginId": true, "name": true, "description": true, "author": true, "version": true},
}

func (s *Store) Search(ctx context.Context, ns, kind string, query entity.Search, page, size int) ([]entity.Record, int, error) {
	if page < 1 || size < 1 || size > 100 || page > 1000000 {
		return nil, 0, fmt.Errorf("%w: invalid pagination", entity.ErrInvalid)
	}
	allowed, ok := searchFields[kind]
	if !ok {
		return nil, 0, entity.ErrInvalid
	}
	args := []any{ns, kind}
	where := []string{"namespace=$1", "kind=$2"}
	bind := func(v any) string { args = append(args, v); return fmt.Sprintf("$%d", len(args)) }
	field := func(name string) (string, error) {
		if name != "" {
			name = strings.ToLower(name[:1]) + name[1:]
		}
		if !allowed[name] {
			return "", fmt.Errorf("%w: unknown search field", entity.ErrInvalid)
		}
		return name, nil
	}
	for _, filter := range query.Filters {
		name, err := field(filter.Key)
		if err != nil {
			return nil, 0, err
		}
		if filter.MatchMode < 1 || filter.MatchMode > 3 {
			return nil, 0, entity.ErrInvalid
		}
		var value any
		if json.Unmarshal(filter.Value, &value) != nil {
			return nil, 0, entity.ErrInvalid
		}
		if name == "tags" {
			if values, ok := value.([]any); ok {
				if filter.MatchMode != 3 {
					return nil, 0, entity.ErrInvalid
				}
				tags := make([]string, 0, len(values))
				for _, v := range values {
					tag, ok := v.(string)
					if !ok {
						return nil, 0, entity.ErrInvalid
					}
					tags = append(tags, tag)
				}
				where = append(where, "(data->'tags') ?| "+bind(tags)+"::text[]")
				continue
			}
		}
		text := fmt.Sprint(value)
		operator := "="
		if filter.MatchMode != 1 {
			operator = "ILIKE"
			text = strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(text)
			if filter.MatchMode == 2 {
				text += "%"
			} else {
				text = "%" + text + "%"
			}
		}
		keyName := name
		if name == "dataSourceId" {
			keyName = "id"
		}
		key, valueArg := bind(keyName), bind(text)
		if name == "tags" {
			where = append(where, "EXISTS(SELECT 1 FROM jsonb_array_elements_text(data->"+key+") AS tag WHERE lower(tag) "+operator+" lower("+valueArg+"))")
		} else {
			expr := "data->>" + key
			if name == "dataSourceId" {
				expr = "data->'dataSource'->>" + key
			}
			where = append(where, "lower("+expr+") "+operator+" lower("+valueArg+")")
		}
	}
	condition := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow(ctx, "SELECT count(*) FROM datacat.resources WHERE "+condition, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	order := "id"
	if query.Sort != nil {
		name, err := field(query.Sort.FieldName)
		if err != nil {
			return nil, 0, err
		}
		order = "(data->>" + bind(name) + ")"
		switch query.Sort.SortOrder {
		case 0, 1:
			order += " ASC"
		case -1:
			order += " DESC"
		default:
			return nil, 0, entity.ErrInvalid
		}
		order += ",id"
	}
	limit, offset := bind(size), bind((page-1)*size)
	rows, err := s.db.Query(ctx, "SELECT namespace,kind,id,name,data,version FROM datacat.resources WHERE "+condition+" ORDER BY "+order+" LIMIT "+limit+" OFFSET "+offset, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]entity.Record, 0, size)
	for rows.Next() {
		var r entity.Record
		if err = rows.Scan(&r.Namespace, &r.Kind, &r.ID, &r.Name, &r.Data, &r.Version); err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}
