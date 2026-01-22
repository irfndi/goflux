package trading

type Position struct {
	entryOrder *Order
	exitOrder  *Order
}

func (p *Position) IsNew() bool {
	return p.entryOrder == nil && p.exitOrder == nil
}

func (p *Position) IsOpen() bool {
	return p.entryOrder != nil && p.exitOrder == nil
}

func (p *Position) EntranceOrder() *Order {
	return p.entryOrder
}

func (p *Position) ExitOrder() *Order {
	return p.exitOrder
}

func (p *Position) Enter(order Order) {
	p.entryOrder = &order
}

func (p *Position) Exit(order Order) {
	p.exitOrder = &order
}
