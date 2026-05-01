package symbol_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/achannarasappa/ticker/v5/internal/cli/symbol"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("SymbolPanic", func() {
	var server *ghttp.Server

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	It("should not panic with malformed CSV missing columns", func() {
		responseFixture := `"SOMESYMBOL.X","some-symbol"
`
		server.RouteToHandler("GET", "/symbols.csv",
			ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, responseFixture, http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}}),
			),
		)

		Expect(func() {
			symbol.GetTickerSymbols(server.URL() + "/symbols.csv")
		}).ToNot(Panic())
	})
})
