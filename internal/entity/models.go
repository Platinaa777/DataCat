package entity

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Duration keeps the TimeSpan format used by existing clients.
type Duration time.Duration

func ParseDuration(s string) (Duration, error) {
	if d, err := time.ParseDuration(s); err == nil && d >= 0 {
		return Duration(d), nil
	}
	var days int64
	if dot := strings.IndexByte(s, '.'); dot >= 0 && dot < strings.IndexByte(s, ':') {
		var err error
		days, err = strconv.ParseInt(s[:dot], 10, 32)
		if err != nil || days < 0 {
			return 0, fmt.Errorf("%w: duration", ErrInvalid)
		}
		s = s[dot+1:]
	}
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("%w: duration", ErrInvalid)
	}
	hours, e1 := strconv.ParseInt(parts[0], 10, 32)
	minutes, e2 := strconv.ParseInt(parts[1], 10, 32)
	seconds, e3 := strconv.ParseFloat(parts[2], 64)
	if e1 != nil || e2 != nil || e3 != nil || hours < 0 || hours > 2400000 || minutes < 0 || minutes > 59 || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds < 0 || seconds >= 60 || days > 100000 {
		return 0, fmt.Errorf("%w: duration", ErrInvalid)
	}
	if float64(days)*86400+float64(hours)*3600+float64(minutes)*60+seconds > float64(math.MaxInt64)/float64(time.Second) {
		return 0, fmt.Errorf("%w: duration overflow", ErrInvalid)
	}
	return Duration(time.Duration(days)*24*time.Hour + time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds*float64(time.Second))), nil
}

func (d Duration) String() string {
	n := time.Duration(d)
	days := n / (24 * time.Hour)
	n %= 24 * time.Hour
	s := fmt.Sprintf("%02d:%02d:%02d", n/time.Hour, n/time.Minute%60, n/time.Second%60)
	if frac := n % time.Second; frac != 0 {
		s += strings.TrimRight(fmt.Sprintf(".%09d", frac), "0")
	}
	if days > 0 {
		s = fmt.Sprintf("%d.%s", days, s)
	}
	return s
}
func (d Duration) MarshalJSON() ([]byte, error) { return json.Marshal(d.String()) }
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("%w: duration must be a string", ErrInvalid)
	}
	v, err := ParseDuration(s)
	if err == nil {
		*d = v
	}
	return err
}

type Namespace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Dashboard struct {
	ID          string    `json:"id"`
	NamespaceID string    `json:"namespaceId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Tags        []string  `json:"tags"`
}
type DataSource struct {
	ID               string `json:"id"`
	NamespaceID      string `json:"namespaceId"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	TypeID           int64  `json:"typeId"`
	Purpose          string `json:"purpose"`
	ConnectionString string `json:"connectionString"`
}
type RawQuery struct {
	Query      string     `json:"query"`
	DataSource DataSource `json:"dataSource"`
}
type Panel struct {
	ID                 string   `json:"id"`
	NamespaceID        string   `json:"namespaceId"`
	Title              string   `json:"title"`
	TypeName           string   `json:"typeName"`
	Query              RawQuery `json:"query"`
	Layout             string   `json:"layout"`
	DashboardID        string   `json:"dashboardId"`
	StyleConfiguration string   `json:"styleConfiguration"`
}
type Variable struct {
	ID          string  `json:"id"`
	Placeholder string  `json:"placeholder"`
	Value       string  `json:"value"`
	NamespaceID string  `json:"namespaceId"`
	DashboardID *string `json:"dashboardId"`
}
type NamedType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type NotificationChannel struct {
	ID              int64  `json:"id"`
	DestinationName string `json:"destinationName"`
	Settings        string `json:"settings"`
	NamespaceID     string `json:"namespaceId"`
	GroupID         string `json:"notificationChannelGroupId"`
}
type NotificationGroup struct {
	ID                   string                `json:"id"`
	Name                 string                `json:"name"`
	NamespaceID          string                `json:"namespaceId"`
	NotificationChannels []NotificationChannel `json:"notificationChannels"`
}
type Alert struct {
	ID                       string            `json:"id"`
	NamespaceID              string            `json:"namespaceId"`
	Description              string            `json:"description"`
	Template                 string            `json:"template"`
	RawQuery                 string            `json:"rawQuery"`
	Status                   string            `json:"status"`
	DataSource               DataSource        `json:"dataSource"`
	NotificationChannelGroup NotificationGroup `json:"notificationChannelGroup"`
	PreviousExecutionTime    time.Time         `json:"previousExecutionTime"`
	NextExecutionTime        time.Time         `json:"nextExecutionTime"`
	WaitTimeBeforeAlerting   Duration          `json:"waitTimeBeforeAlerting"`
	RepeatInterval           Duration          `json:"repeatInterval"`
	Tags                     []string          `json:"tags"`
	PendingSince             *time.Time        `json:"pendingSince,omitempty"`
	LastNotifiedAt           *time.Time        `json:"lastNotifiedAt,omitempty"`
}
type Plugin struct {
	ID          string `json:"pluginId"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Settings    string `json:"settings"`
	Endpoint    string `json:"endpoint"`
	IsActive    bool   `json:"isActive"`
}
type RoleMapping struct {
	ExternalRole string `json:"externalRole"`
	NamespaceID  string `json:"namespaceId"`
	RoleID       int    `json:"roleId"`
}
type Claim struct {
	Role        string `json:"role"`
	NamespaceID string `json:"namespaceId"`
}
type Principal struct {
	Subject     string
	Claims      []Claim
	GlobalAdmin bool
}

func (p Principal) Can(namespace string, write bool) bool {
	if p.GlobalAdmin {
		return true
	}
	for _, c := range p.Claims {
		if c.NamespaceID == namespace && (c.Role == "Admin" || c.Role == "Editor" || (!write && c.Role == "Viewer")) {
			return true
		}
	}
	return false
}

type Record struct {
	Namespace string
	Kind      string
	ID        string
	Name      string
	Data      json.RawMessage
	Version   int64
}

type SearchFilter struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	MatchMode MatchMode       `json:"matchMode"`
	FieldType SearchFieldType `json:"fieldType"`
}
type Sort struct {
	FieldName string    `json:"fieldName"`
	SortOrder SortOrder `json:"sortOrder"`
}
type Search struct {
	Filters []SearchFilter `json:"filters"`
	Sort    *Sort          `json:"sort"`
}
type Page[T any] struct {
	Items           []T  `json:"items"`
	TotalCount      int  `json:"totalCount"`
	PageNumber      int  `json:"pageNumber"`
	PageSize        int  `json:"pageSize"`
	TotalPages      int  `json:"totalPages"`
	HasPreviousPage bool `json:"hasPreviousPage"`
	HasNextPage     bool `json:"hasNextPage"`
}

func NewPage[T any](items []T, total, number, size int) Page[T] {
	pages := (total + size - 1) / size
	return Page[T]{items, total, number, size, pages, number > 1, number < pages}
}
