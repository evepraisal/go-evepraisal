package parsers

type Parser func(input Input) (ParserResult, Input)

type ParserResult interface {
	// Name is the name of the PARSER that yielded this result
	Name() string
	Lines() []int
}

var AllParsers = []Parser{
	ParseKillmail,
	ParseEFT,
	ParseFitting,
	ParseLootHistory,
	ParsePI,
	ParseViewContents,
	ParseMiningLedger,
	ParseWallet,
	ParseSurveyScan,
	ParseContract,
	ParseAssets,
	ParseIndustry,
	ParseCargoScan,
	ParseDScan,
	ParseListing,
}
