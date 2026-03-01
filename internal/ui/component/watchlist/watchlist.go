package watchlist

import (
	"fmt"
	"strings"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	s "github.com/achannarasappa/ticker/v5/internal/sorter"
	row "github.com/achannarasappa/ticker/v5/internal/ui/component/watchlist/row"
	u "github.com/achannarasappa/ticker/v5/internal/ui/util"

	tea "github.com/charmbracelet/bubbletea"
)

// Config represents the configuration for the watchlist component
type Config struct {
	Separate              bool
	ShowPositions         bool
	ExtraInfoExchange     bool
	ExtraInfoFundamentals bool
	Sort                  string
	Styles                c.Styles
}

// Model for watchlist section
type Model struct {
	width          int
	assets         []*c.Asset
	assetsBySymbol map[string]*c.Asset
	sorter         s.Sorter
	config         Config
	cellWidths     row.CellWidthsContainer
	rows           []*row.Model
	rowsBySymbol   map[string]*row.Model
	rowsByID       map[int]int
}

// Messages for replacing assets
type SetAssetsMsg []c.Asset

// Messages for updating assets
type UpdateAssetsMsg []c.Asset

// Messages for changing sort
type ChangeSortMsg string

// NewModel returns a model with default values
func NewModel(config Config) *Model {
	return &Model{
		width:          80,
		config:         config,
		assets:         make([]*c.Asset, 0),
		assetsBySymbol: make(map[string]*c.Asset),
		sorter:         s.NewSorter(config.Sort),
		rowsBySymbol:   make(map[string]*row.Model),
		rowsByID:       make(map[int]int),
	}
}

// Init initializes the watchlist
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the watchlist
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SetAssetsMsg:

		var cmd tea.Cmd
		cmds := make([]tea.Cmd, 0)

		// Convert []c.Asset to []*c.Asset and update assetsBySymbol map
		assets := make([]*c.Asset, len(msg))
		assetsBySymbol := make(map[string]*c.Asset)

		for i := range msg {
			assets[i] = &msg[i]
			assetsBySymbol[msg[i].Symbol] = assets[i]
		}

		assets = m.sorter(assets)

		rowsByID := make(map[int]int)

		for i, asset := range assets {
			if i < len(m.rows) {
				m.rows[i], cmd = m.rows[i].Update(row.UpdateAssetMsg(asset))
				cmds = append(cmds, cmd)
				m.rowsBySymbol[assets[i].Symbol] = m.rows[i]
			} else {
				m.rows = append(m.rows, row.New(row.Config{
					Separate:              m.config.Separate,
					ExtraInfoExchange:     m.config.ExtraInfoExchange,
					ExtraInfoFundamentals: m.config.ExtraInfoFundamentals,
					ShowPositions:         m.config.ShowPositions,
					Styles:                m.config.Styles,
					Asset:                 asset,
				}))
				m.rowsBySymbol[assets[i].Symbol] = m.rows[len(m.rows)-1]
			}
			rowsByID[m.rows[i].ID()] = i
		}

		if len(assets) < len(m.rows) {
			m.rows = m.rows[:len(assets)]
		}

		m.assets = assets
		m.assetsBySymbol = assetsBySymbol
		m.rowsByID = rowsByID

		// TODO: only set conditionally if all assets have changed
		m.cellWidths = getCellWidths(m.assets)
		for i, r := range m.rows {
			m.rows[i], _ = r.Update(row.SetCellWidthsMsg{
				Width:      m.width,
				CellWidths: m.cellWidths,
			})
		}

		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:

		m.width = msg.Width
		m.cellWidths = getCellWidths(m.assets)
		for i, r := range m.rows {
			m.rows[i], _ = r.Update(row.SetCellWidthsMsg{
				Width:      m.width,
				CellWidths: m.cellWidths,
			})
		}

		return m, nil

	case row.FrameMsg:

		var cmd tea.Cmd

		if i, ok := m.rowsByID[int(msg)]; ok {
			m.rows[i], cmd = m.rows[i].Update(msg)

			return m, cmd
		}

		return m, nil

	case ChangeSortMsg:

		var cmd tea.Cmd
		cmds := make([]tea.Cmd, 0)

		// Update the sorter with the new sort option
		m.config.Sort = string(msg)
		m.sorter = s.NewSorter(m.config.Sort)

		// Re-sort and update the assets
		assets := m.sorter(m.assets)
		m.assets = assets

		rowsByID := make(map[int]int)

		// Update rows with the new order
		for i, asset := range assets {
			m.rows[i], cmd = m.rows[i].Update(row.UpdateAssetMsg(asset))
			cmds = append(cmds, cmd)
			rowsByID[m.rows[i].ID()] = i
		}

		m.rowsByID = rowsByID

		return m, tea.Batch(cmds...)

	}

	return m, nil
}

// View rendering hook for bubbletea
func (m *Model) View() string {

	if m.width < 80 {
		return fmt.Sprintf("Terminal window too narrow to render content\nResize to fix (%d/80)", m.width)
	}

	rows := make([]string, 0)
	for _, row := range m.rows {
		rows = append(rows, row.View())
	}

	return strings.Join(rows, "\n")

}
func getCellWidths(assets []*c.Asset) row.CellWidthsContainer {

	cellMaxWidths := row.CellWidthsContainer{}

	for _, asset := range assets {
		var quoteLength int

		volumeMarketCapLength := len(u.ConvertFloatToString(asset.QuoteExtended.MarketCap, true))

		if asset.QuoteExtended.FiftyTwoWeekHigh == 0.0 {
			quoteLength = len(u.ConvertFloatToString(asset.QuotePrice.Price, asset.Meta.IsVariablePrecision))
		}

		if asset.QuoteExtended.FiftyTwoWeekHigh != 0.0 {
			quoteLength = len(u.ConvertFloatToString(asset.QuoteExtended.FiftyTwoWeekHigh, asset.Meta.IsVariablePrecision))
		}

		if volumeMarketCapLength > cellMaxWidths.WidthVolumeMarketCap {
			cellMaxWidths.WidthVolumeMarketCap = volumeMarketCapLength
		}

		if quoteLength > cellMaxWidths.QuoteLength {
			cellMaxWidths.QuoteLength = quoteLength
			cellMaxWidths.WidthQuote = quoteLength + row.WidthChangeStatic
			cellMaxWidths.WidthQuoteExtended = quoteLength
			cellMaxWidths.WidthQuoteRange = row.WidthRangeStatic + (quoteLength * 2)
		}

		if asset.Position != (c.Position{}) {
			positionLength := len(u.ConvertFloatToString(asset.Position.Value, asset.Meta.IsVariablePrecision))
			positionQuantityLength := len(u.ConvertFloatToString(asset.Position.Quantity, asset.Meta.IsVariablePrecision))
			positionUnitCostLength := len(u.ConvertFloatToString(asset.Position.UnitCost, asset.Meta.IsVariablePrecision))

			if positionLength > cellMaxWidths.PositionLength {
				cellMaxWidths.PositionLength = positionLength
				cellMaxWidths.WidthPosition = positionLength + row.WidthChangeStatic + row.WidthPositionGutter
			}

			if positionLength > cellMaxWidths.WidthPositionExtended {
				cellMaxWidths.WidthPositionExtended = positionLength
			}

			if positionQuantityLength > cellMaxWidths.WidthPositionExtended {
				cellMaxWidths.WidthPositionExtended = positionQuantityLength
			}

			if positionUnitCostLength > cellMaxWidths.WidthPositionExtended {
				cellMaxWidths.WidthPositionExtended = positionUnitCostLength
			}

		}

	}

	return cellMaxWidths

}
