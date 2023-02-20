package oracle

import (
	"fmt"
	"net/url"
	"sync"

	oracletypes "github.com/shentufoundation/shentu/v2/x/oracle/types"

	"github.com/shentufoundation/oracle-operator/types"
)

func getPrimitivePayload(msg oracletypes.MsgCreateTask) (types.PrimitivePayload, error) {
	client, contract, err := parseMsgCreateTaskContract(msg.Contract)
	if err != nil {
		return types.PrimitivePayload{}, fmt.Errorf("wrong task contract format: %s", err.Error())
	}
	return types.PrimitivePayload{Client: client, Address: contract, Function: msg.Function, Contract: msg.Contract}, nil
}

// queryPrimitive gets score for each primitive.
func queryPrimitive(
	ctx types.Context,
	primitive types.Primitive,
	payload types.PrimitivePayload,
	primitiveScores chan<- types.PrimitiveScore,
	wg *sync.WaitGroup,
) {
	logger := ctx.Logger()
	logger.Debug("query primitive", "contract", payload.Contract, "function", payload.Function)

	endpoint := url.URL{
		Scheme: "https",
		Host:   primitive.PrimitiveType + ".certik-skynet.com",
		Path:   "score",
	}
	q := endpoint.Query()
	q.Set("address", payload.Contract)
	endpoint.RawQuery = q.Encode()
	endpointUrl := endpoint.String()

	score, err := handleRequest(
		ctx.WithLoggerLabels("submodule", "querier", "endpoint", endpointUrl, "payload", payload),
		endpointUrl,
		payload,
	)
	if err != nil {
		logger.Error(err.Error())
		wg.Done()
		return
	}

	logger.Debug("got score from primitive endpoint", "url", endpointUrl, "score", score)
	primitiveScores <- types.PrimitiveScore{Score: score, Primitive: primitive}
	wg.Done()
}
