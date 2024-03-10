package titletags

const CEO = "CEO" // Chief executive officer

// Development team titles
const CTO = "CTO"            // Chief technology officer
const QA = "QA"              // Quality Assurance
const AQA = "AQA"            // Automation Quality Assurance
const DEV_MGR = "DevManager" // Developer manager
const DEV_BE = "BE"          // Developer FrontEnd
const DEV_FE = "FE"          // Developer BackEnd
const DES_UI = "UI"          // User Interface Designer
const DES_UX = "UX"          // User Experience Designer
const PO = "PO"              // Product Owner

// Marketing team titles
// Help here

// All titles, If a new title is added make sure to add it here
var TITLES = []string{
	CTO, AQA, QA, DEV_MGR, DEV_BE, DEV_FE, DES_UI, DES_UX, PO,
}

// Titles to always be includeded in meeting order
var ALWAYS_INCLUDE = []string{
	QA,
}

// Map of titles that when called out for meeting will also pull from other roles
var TITLE_INCLUSION = map[string][]string{
	DEV_FE: {DES_UI, DES_UX, AQA},
}
