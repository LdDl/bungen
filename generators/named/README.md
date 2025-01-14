## Basic model generator with named structures

Author: [@Dionid](https://github.com/Dionid)

Use `model-named` sub-command to execute generator:

`bungen model-named -h`

First create your database and tables in it

```bun
create table "projects"
(
    "projectId" serial not null,
    "name"      text   not null,

    primary key ("projectId")
);

create table "users"
(
    "userId"    serial      not null,
    "email"     varchar(64) not null,
    "activated" bool        not null default false,
    "name"      varchar(128),
    "countryId" integer,

    primary key ("userId")
);

create schema "geo";
create table geo."countries"
(
    "countryId" serial     not null,
    "code"      varchar(3) not null,
    "coords"    integer[],

    primary key ("countryId")
);

alter table "users"
    add constraint "fk_user_country"
        foreign key ("countryId")
            references geo."countries" ("countryId") on update restrict on delete restrict;
```

### Run generator

`bungen model-named -c postgres://user:password@localhost:5432/yourdb -o ~/output/model.go -t public.* -f`

You should get following models on model package:

```go
//lint:file-ignore U1000 ignore unused code, it's generated
package model

type ColumnsProject struct {
	ID, Name string
}

type ColumnsUser struct {
	ID, Email, Activated, Name, CountryID string
	Country                               string
}

type ColumnsGeoCountry struct {
	ID, Code, Coords string
}

type ColumnsSt struct {
	Project    ColumnsProject
	User       ColumnsUser
	GeoCountry ColumnsGeoCountry
}

var Columns = ColumnsSt{
	Project: ColumnsProject{
		ID:   "projectId",
		Name: "name",
	},
	User: ColumnsUser{
		ID:        "userId",
		Email:     "email",
		Activated: "activated",
		Name:      "name",
		CountryID: "countryId",

		Country: "Country",
	},
	GeoCountry: ColumnsGeoCountry{
		ID:     "countryId",
		Code:   "code",
		Coords: "coords",
	},
}

type TableProject struct {
	Name, Alias string
}

type TableUser struct {
	Name, Alias string
}

type TableGeoCountry struct {
	Name, Alias string
}

type TablesSt struct {
	Project    TableProject
	User       TableUser
	GeoCountry TableGeoCountry
}

var Tables = TablesSt{
	Project: TableProject{
		Name:  "projects",
		Alias: "t",
	},
	User: TableUser{
		Name:  "users",
		Alias: "t",
	},
	GeoCountry: TableGeoCountry{
		Name:  "geo.countries",
		Alias: "t",
	},
}

type Project struct {
	bun.BaseModel `bun:"projects,alias:t"`

	ID   int    `bun:"projectId,pk"`
	Name string `bun:"name,nullzero"`
}

type User struct {
	bun.BaseModel `bun:"users,alias:t"`

	ID        int     `bun:"userId,pk"`
	Email     string  `bun:"email,nullzero"`
	Activated bool    `bun:"activated,nullzero"`
	Name      *string `bun:"name"`
	CountryID *int    `bun:"countryId"`

	Country *GeoCountry `bun:"join:countryId=countryId,rel:belongs-to"`
}

type GeoCountry struct {
	bun.BaseModel `bun:"geo.countries,alias:t"`

	ID     int    `bun:"countryId,pk"`
	Code   string `bun:"code,nullzero"`
	Coords []int  `bun:"coords,array"`
}

```

### Try it

```go
package model

import (
	"fmt"
	"testing"

	"github.com/uptrace/bun"
)

const AllColumns = "t.*"

func TestModel(t *testing.T) {
	// connecting to db
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN("postgres://user:password@localhost:5432/yourdb")))
	db := bun.NewDB(pgdb, pgdialect.New(), bun.WithDiscardUnknownColumns())

	if _, err := db.Exec(`truncate table users; truncate table geo.countries cascade;`); err != nil {
		panic(err)
	}

	// objects to insert
	toInsert := []GeoCountry{
		GeoCountry{
			Code:   "us",
			Coords: []int{1, 2},
		},
		GeoCountry{
			Code:   "uk",
			Coords: nil,
		},
	}

	// inserting
	ctx := context.Background()
	if _, err := db.NewInsert().Model(&toInsert).Column("code", "coords").Exec(ctx); err != nil {
		panic(err)
	}

	// selecting
	var toSelect []GeoCountry
	ctx = context.Background()
	if err := db.NewSelect().Model(&toSelect).Scan(ctx); err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", toSelect)

	// user with fk
	newUser := User{
		Email:     "test@gmail.com",
		Activated: true,
		CountryID: &toSelect[0].ID,
	}

	// inserting
	ctx = context.Background()
	if _, err := db.NewInsert().Model(&newUser).Column("email", "activated", "countryId").Exec(ctx); err != nil {
		panic(err)
	}

	// selecting inserted user
	user := User{}
	m := db.NewSelect().
		Column(AllColumns).
		Relation(Columns.User.Country).
		Where(`? = ?`, bun.Ident(Columns.User.Email), "test@gmail.com")

	ctx = context.Background()
	if err := m.Scan(ctx); err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", user)
	fmt.Printf("%#v\n", user.Country)
}

```
