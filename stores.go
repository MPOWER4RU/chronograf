package chronograf

import (
	"time"

	"golang.org/x/net/context"
)

// UserID is a unique ID for a source user.
type UserID int

// Represents an authenticated user.
type User struct {
	ID   UserID
	Name string
}

// AuthStore is the Storage and retrieval of authentication information
type AuthStore struct {
	// User management for the AuthStore
	Users interface {
		// Create a new User in the AuthStore
		Add(context.Context, User) error
		// Delete the User from the AuthStore
		Delete(context.Context, User) error
		// Retrieve a user if `ID` exists.
		Get(ctx context.Context, ID int) error
		// Update the user's permissions or roles
		Update(context.Context, User) error
	}
}

// ExplorationID is a unique ID for an Exploration.
type ExplorationID int

// Exploration is a serialization of front-end Data Explorer.
type Exploration struct {
	ID        ExplorationID
	Name      string    // User provided name of the Exploration.
	UserID    UserID    // UserID is the owner of this Exploration.
	Data      string    // Opaque blob of JSON data.
	CreatedAt time.Time // Time the exploration was first created.
	UpdatedAt time.Time // Latest time the exploration was updated.
	Default   bool      // Flags an exploration as the default.
}

// ExplorationStore stores front-end serializations of data explorer sessions.
type ExplorationStore interface {
	// Search the ExplorationStore for each Exploration owned by `UserID`.
	Query(ctx context.Context, userID UserID) ([]*Exploration, error)
	// Create a new Exploration in the ExplorationStore.
	Add(context.Context, *Exploration) (*Exploration, error)
	// Delete the Exploration from the ExplorationStore.
	Delete(context.Context, *Exploration) error
	// Retrieve an Exploration if `ID` exists.
	Get(ctx context.Context, ID ExplorationID) (*Exploration, error)
	// Update the Exploration; will also update the `UpdatedAt` time.
	Update(context.Context, *Exploration) error
}

// Cell is a rectangle and multiple time series queries to visualize.
type Cell struct {
	X       int32   `json:"x"`
	Y       int32   `json:"y"`
	W       int32   `json:"w"`
	H       int32   `json:"h"`
	Queries []Query `json:"queries"`
}

// Layout is a collection of Cells for visualization
type Layout struct {
	ID          string `json:"id"`
	Application string `json:"app"`
	Measurement string `json:"measurement"`
	Cells       []Cell `json:"cells"`
}

// LayoutStore stores dashboards and associated Cells
type LayoutStore interface {
	// All returns all dashboards in the store
	All(context.Context) ([]Layout, error)
	// Add creates a new dashboard in the LayoutStore
	Add(context.Context, Layout) (Layout, error)
	// Delete the dashboard from the store
	Delete(context.Context, Layout) error
	// Get retrieves Layout if `ID` exists
	Get(ctx context.Context, ID string) (Layout, error)
	// Update the dashboard in the store.
	Update(context.Context, Layout) error
}

type Source struct {
	ID       int      // ID is the unique ID of the source
	Name     string   // Name is the user-defined name for the source
	Type     string   // Type specifies which kinds of source (enterprise vs oss)
	Username string   // Username is the username to connect to the source
	Password string   // Password is in CLEARTEXT FIXME
	URL      []string // URL are the connections to the source
	Default  bool     // Default specifies the default source for the application
}

// SourcesStore stores connection information for a `TimeSeries`
type SourcesStore interface {
	// All returns all sources in the store
	All(context.Context) ([]Source, error)
	// Add creates a new source in the SourcesStore and returns Source with ID
	Add(context.Context, Source) (Source, error)
	// Delete the Source from the store
	Delete(context.Context, Source) error
	// Get retrieves Source if `ID` exists
	Get(ctx context.Context, ID int) (Source, error)
	// Update the Source in the store.
	Update(context.Context, Source) error
}

type Server struct {
	ID       int    // ID is the unique ID of the server
	SrcID    int    // SrcID of the data soruce
	Name     string // Name is the user-defined name for the server
	Username string // Username is the username to connect to the server
	Password string // Password is in CLEARTEXT FIXME
	URL      string // URL are the connections to the server
}

// ServersStore stores connection information for a `Server`
type ServersStore interface {
	// All returns all servers in the store
	All(context.Context) ([]Server, error)
	// Add creates a new source in the ServersStore and returns Server with ID
	Add(context.Context, Server) (Server, error)
	// Delete the Server from the store
	Delete(context.Context, Server) error
	// Get retrieves Server if `ID` exists
	Get(ctx context.Context, ID int) (Server, error)
	// Update the Server in the store.
	Update(context.Context, Server) error
}
