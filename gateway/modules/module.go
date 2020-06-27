package modules

import (
	"github.com/spaceuptech/space-cloud/gateway/model"
	"github.com/spaceuptech/space-cloud/gateway/modules/auth"
	"github.com/spaceuptech/space-cloud/gateway/modules/crud"
	"github.com/spaceuptech/space-cloud/gateway/modules/crud/driver"
	"github.com/spaceuptech/space-cloud/gateway/modules/eventing"
	"github.com/spaceuptech/space-cloud/gateway/modules/filestore"
	"github.com/spaceuptech/space-cloud/gateway/modules/functions"
	"github.com/spaceuptech/space-cloud/gateway/modules/realtime"
	"github.com/spaceuptech/space-cloud/gateway/modules/schema"
	"github.com/spaceuptech/space-cloud/gateway/modules/userman"
	"github.com/spaceuptech/space-cloud/gateway/utils/admin"
	"github.com/spaceuptech/space-cloud/gateway/utils/graphql"
	"github.com/spaceuptech/space-cloud/gateway/utils/metrics"
	"github.com/spaceuptech/space-cloud/gateway/utils/syncman"
)

// Modules is an object that sets up the modules
type Module struct {
	auth      *auth.Module
	db        *crud.Module
	user      *userman.Module
	file      *filestore.Module
	functions *functions.Module
	realtime  *realtime.Module
	eventing  *eventing.Module
	graphql   *graphql.Module
	schema    *schema.Schema
}

func newModule(nodeID string, syncMan *syncman.Manager, adminMan *admin.Manager, metrics *metrics.Module, driver *driver.Handler) *Module {
	c := crud.Init()
	c.SetAdminManager(adminMan)
	c.SetGetSecrets(syncMan.GetSecrets)

	s := schema.Init(c)
	c.SetSchema(s)

	a := auth.Init(nodeID, c)
	a.SetMakeHTTPRequest(syncMan.MakeHTTPRequest)

	fn := functions.Init(a, syncMan, metrics.AddFunctionOperation)
	f := filestore.Init(a, metrics.AddFileOperation)
	f.SetGetSecrets(syncMan.GetSecrets)

	e := eventing.New(a, c, s, adminMan, syncMan, f, metrics.AddEventingType)
	f.SetEventingModule(e)

	c.SetHooks(&model.CrudHooks{
		Create: e.HookDBCreateIntent,
		Update: e.HookDBUpdateIntent,
		Delete: e.HookDBDeleteIntent,
		Batch:  e.HookDBBatchIntent,
		Stage:  e.HookStage,
	}, metrics.AddDBOperation)

	rt, _ := realtime.Init(nodeID, e, a, c, s, metrics, syncMan)

	u := userman.Init(c, a)
	graphqlMan := graphql.New(a, c, fn, s)

	return &Module{auth: a, db: c, user: u, file: f, functions: fn, realtime: rt, eventing: e, graphql: graphqlMan, schema: s}
}