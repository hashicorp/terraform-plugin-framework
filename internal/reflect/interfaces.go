package reflect

type setUnknownable interface {
	SetUnknown(bool) error
}

type setNullable interface {
	SetNull(bool) error
}
