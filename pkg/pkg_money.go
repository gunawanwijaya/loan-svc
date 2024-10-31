package pkg

import (
	"context"
	"fmt"
	"math"
	"time"

	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type ValidateMoneyError struct {
	UnknownISO4217  string
	RawValue, Value float64
}

func (x *ValidateMoneyError) Error() string {
	switch {
	default:
		return ""
	case x.UnknownISO4217 != "":
		return fmt.Sprintf("money: validate: unknown ISO4217 code for %s", x.UnknownISO4217)
	case x.RawValue != x.Value:
		return fmt.Sprintf("money: validate: different precision value: [%f] from raw value: [%f]", x.Value, x.RawValue)
	}
}

type TakeMoneyError struct {
	Portion float64
}

func (x *TakeMoneyError) Error() string {
	switch {
	default:
		return ""
	case x.Portion < 0 || x.Portion > 1:
		return fmt.Sprintf("money: take: invalid portion [%f] should be ranged between 0 & 1", x.Portion)
	}
}

type Money struct {
	ISO4217 string    `json:"iso4217"` // ISO 4217 code for the representation of currencies - https://en.wikipedia.org/wiki/ISO_4217
	Amount  float64   `json:"amount"`
	Time    time.Time `json:"time"`
	Details string    `json:"details"`
}

func (x *Money) Validate(ctx context.Context) (_ *Money, err error) {
	l, ok := moneyLookup[x.ISO4217]
	if !ok {
		return nil, &ValidateMoneyError{UnknownISO4217: x.ISO4217}
	}

	amount := x.Amount
	if l.Precision == 0 {
		amount = math.Round(x.Amount)
	} else {
		ratio := math.Pow(10, float64(l.Precision))
		amount = math.Round(x.Amount*ratio) / ratio
	}

	if amount != x.Amount {
		err = &ValidateMoneyError{RawValue: x.Amount, Value: amount}
	}
	return &Money{ISO4217: x.ISO4217, Amount: amount, Time: x.Time, Details: x.Details}, err
}

func (x *Money) String() string {
	u, _ := currency.ParseISO(x.ISO4217)
	p := message.NewPrinter(language.English)
	s := currency.NarrowSymbol(u.Amount(x.Amount))
	return p.Sprintf("%f", s)
}

func (x *Money) Sum(y ...*Money) (s *Money, err error) {
	s = &Money{}
	for _, each := range y {
		if each == nil {
			continue
		}
		if x.ISO4217 != each.ISO4217 {
			return x, fmt.Errorf("money: sum: different currency")
		}
		s.Amount += each.Amount
	}
	s.Details = x.Details
	s.ISO4217 = x.ISO4217
	s.Time = x.Time
	s.Amount += x.Amount
	return s, nil
}

func (x *Money) Take(portion float64) (take, remainder *Money, err error) {
	if portion < 0 || portion > 1 {
		return nil, nil, &TakeMoneyError{Portion: portion}
	}
	ctx := context.Background()

	take = &Money{ISO4217: x.ISO4217, Time: x.Time, Details: x.Details, Amount: x.Amount * portion}
	take, _ = take.Validate(ctx) // error validation should be ignored

	remainder = &Money{ISO4217: x.ISO4217, Time: x.Time, Details: x.Details, Amount: x.Amount - take.Amount}
	return take, remainder, err
}

var moneyLookup = map[string]struct {
	NumericalCode string
	Precision     int8
	CurrencyName  string
	Locations     []string
}{
	"AED": {"784", 2, "United Arab Emirates dirham", []string{"United Arab Emirates"}},
	"AFN": {"971", 2, "Afghan afghani", []string{"Afghanistan"}},
	"ALL": {"008", 2, "Albanian lek", []string{"Albania"}},
	"AMD": {"051", 2, "Armenian dram", []string{"Armenia"}},
	"ANG": {"532", 2, "Netherlands Antillean guilder", []string{"Curaçao (CW)", "Sint Maarten (SX)"}},
	"AOA": {"973", 2, "Angolan kwanza", []string{"Angola"}},
	"ARS": {"032", 2, "Argentine peso", []string{"Argentina"}},
	"AUD": {"036", 2, "Australian dollar", []string{"Australia", "Christmas Island (CX)", "Cocos (Keeling) Islands (CC)", "Heard Island and McDonald Islands (HM)", "Kiribati (KI)", "Nauru (NR)", "Norfolk Island (NF)", "Tuvalu (TV)"}},
	"AWG": {"533", 2, "Aruban florin", []string{"Aruba"}},
	"AZN": {"944", 2, "Azerbaijani manat", []string{"Azerbaijan"}},
	"BAM": {"977", 2, "Bosnia and Herzegovina convertible mark", []string{"Bosnia and Herzegovina"}},
	"BBD": {"052", 2, "Barbados dollar", []string{"Barbados"}},
	"BDT": {"050", 2, "Bangladeshi taka", []string{"Bangladesh"}},
	"BGN": {"975", 2, "Bulgarian lev", []string{"Bulgaria"}},
	"BHD": {"048", 3, "Bahraini dinar", []string{"Bahrain"}},
	"BIF": {"108", 0, "Burundian franc", []string{"Burundi"}},
	"BMD": {"060", 2, "Bermudian dollar", []string{"Bermuda"}},
	"BND": {"096", 2, "Brunei dollar", []string{"Brunei Darussalam"}},
	"BOB": {"068", 2, "Boliviano", []string{"Bolivia"}},
	"BOV": {"984", 2, "Bolivian Mvdol (funds code)", []string{"Bolivia"}},
	"BRL": {"986", 2, "Brazilian real", []string{"Brazil"}},
	"BSD": {"044", 2, "Bahamian dollar", []string{"Bahamas"}},
	"BTN": {"064", 2, "Bhutanese ngultrum", []string{"Bhutan"}},
	"BWP": {"072", 2, "Botswana pula", []string{"Botswana"}},
	"BYN": {"933", 2, "Belarusian ruble", []string{"Belarus"}},
	"BZD": {"084", 2, "Belize dollar", []string{"Belize"}},
	"CAD": {"124", 2, "Canadian dollar", []string{"Canada"}},
	"CDF": {"976", 2, "Congolese franc", []string{"Democratic Republic of the Congo"}},
	"CHE": {"947", 2, "WIR euro (complementary currency)", []string{" Switzerland"}},
	"CHF": {"756", 2, "Swiss franc", []string{" Switzerland", "Liechtenstein (LI)"}},
	"CHW": {"948", 2, "WIR franc (complementary currency)", []string{" Switzerland"}},
	"CLF": {"990", 4, "Unidad de Fomento (funds code)", []string{"Chile"}},
	"CLP": {"152", 0, "Chilean peso", []string{"Chile"}},
	"CNY": {"156", 2, "Renminbi[6]", []string{"China"}},
	"COP": {"170", 2, "Colombian peso", []string{"Colombia"}},
	"COU": {"970", 2, "Unidad de Valor Real (UVR) (funds code)[7]", []string{"Colombia"}},
	"CRC": {"188", 2, "Costa Rican colon", []string{"Costa Rica"}},
	"CUP": {"192", 2, "Cuban peso", []string{"Cuba"}},
	"CVE": {"132", 2, "Cape Verdean escudo", []string{"Cabo Verde"}},
	"CZK": {"203", 2, "Czech koruna", []string{"Czechia[8]"}},
	"DJF": {"262", 0, "Djiboutian franc", []string{"Djibouti"}},
	"DKK": {"208", 2, "Danish krone", []string{"Denmark", "Faroe Islands (FO)", "Greenland (GL)"}},
	"DOP": {"214", 2, "Dominican peso", []string{"Dominican Republic"}},
	"DZD": {"012", 2, "Algerian dinar", []string{"Algeria"}},
	"EGP": {"818", 2, "Egyptian pound", []string{"Egypt"}},
	"ERN": {"232", 2, "Eritrean nakfa", []string{"Eritrea"}},
	"ETB": {"230", 2, "Ethiopian birr", []string{"Ethiopia"}},
	"EUR": {"978", 2, "Euro", []string{"Åland Islands (AX)", "Andorra (AD)[c]", "Austria (AT)", "Belgium (BE)", "Croatia (HR)", "Cyprus (CY)", "Estonia (EE)", "European Union (EU)", "Finland (FI)", "France (FR)", "French Guiana (GF)", "French Southern and Antarctic Lands (TF)", "Germany (DE)", "Greece (GR)", "Guadeloupe (GP)", "Ireland (IE)", "Italy (IT)", "Kosovo (XK)[d]", "Latvia (LV)", "Lithuania (LT)", "Luxembourg (LU)", "Malta (MT)", "Martinique (MQ)", "Mayotte (YT)", "Monaco (MC)[c]", "Montenegro (ME)[d]", "Netherlands (NL)", "Portugal (PT)", "Réunion (RE)", "Saint Barthélemy (BL)", "Saint Martin (MF)", "Saint Pierre and Miquelon (PM)", "San Marino (SM)[c]", "Slovakia (SK)", "Slovenia (SI)", "Spain (ES)", "Vatican City (VA)[c]"}},
	"FJD": {"242", 2, "Fiji dollar", []string{"Fiji"}},
	"FKP": {"238", 2, "Falkland Islands pound", []string{"Falkland Islands (pegged to GBP 1:1)"}},
	"GBP": {"826", 2, "Pound sterling", []string{"United Kingdom", "Isle of Man (IM, see Manx pound)", "Jersey (JE, see Jersey pound)", "Guernsey (GG, see Guernsey pound)", "Tristan da Cunha (SH-TA)"}},
	"GEL": {"981", 2, "Georgian lari", []string{"Georgia"}},
	"GHS": {"936", 2, "Ghanaian cedi", []string{"Ghana"}},
	"GIP": {"292", 2, "Gibraltar pound", []string{"Gibraltar (pegged to GBP 1:1)"}},
	"GMD": {"270", 2, "Gambian dalasi", []string{"Gambia"}},
	"GNF": {"324", 0, "Guinean franc", []string{"Guinea"}},
	"GTQ": {"320", 2, "Guatemalan quetzal", []string{"Guatemala"}},
	"GYD": {"328", 2, "Guyanese dollar", []string{"Guyana"}},
	"HKD": {"344", 2, "Hong Kong dollar", []string{"Hong Kong"}},
	"HNL": {"340", 2, "Honduran lempira", []string{"Honduras"}},
	"HTG": {"332", 2, "Haitian gourde", []string{"Haiti"}},
	"HUF": {"348", 2, "Hungarian forint", []string{"Hungary"}},
	"IDR": {"360", 2, "Indonesian rupiah", []string{"Indonesia"}},
	"ILS": {"376", 2, "Israeli new shekel", []string{"Israel"}},
	"INR": {"356", 2, "Indian rupee", []string{"India", "Bhutan"}},
	"IQD": {"368", 3, "Iraqi dinar", []string{"Iraq"}},
	"IRR": {"364", 2, "Iranian rial", []string{"Iran"}},
	"ISK": {"352", 0, "Icelandic króna (plural: krónur)", []string{"Iceland"}},
	"JMD": {"388", 2, "Jamaican dollar", []string{"Jamaica"}},
	"JOD": {"400", 3, "Jordanian dinar", []string{"Jordan"}},
	"JPY": {"392", 0, "Japanese yen", []string{"Japan"}},
	"KES": {"404", 2, "Kenyan shilling", []string{"Kenya"}},
	"KGS": {"417", 2, "Kyrgyzstani som", []string{"Kyrgyzstan"}},
	"KHR": {"116", 2, "Cambodian riel", []string{"Cambodia"}},
	"KMF": {"174", 0, "Comoro franc", []string{"Comoros"}},
	"KPW": {"408", 2, "North Korean won", []string{"North Korea"}},
	"KRW": {"410", 0, "South Korean won", []string{"South Korea"}},
	"KWD": {"414", 3, "Kuwaiti dinar", []string{"Kuwait"}},
	"KYD": {"136", 2, "Cayman Islands dollar", []string{"Cayman Islands"}},
	"KZT": {"398", 2, "Kazakhstani tenge", []string{"Kazakhstan"}},
	"LAK": {"418", 2, "Lao kip", []string{"Lao People's Democratic Republic"}},
	"LBP": {"422", 2, "Lebanese pound", []string{"Lebanon"}},
	"LKR": {"144", 2, "Sri Lankan rupee", []string{"Sri Lanka"}},
	"LRD": {"430", 2, "Liberian dollar", []string{"Liberia"}},
	"LSL": {"426", 2, "Lesotho loti", []string{"Lesotho"}},
	"LYD": {"434", 3, "Libyan dinar", []string{"Libya"}},
	"MAD": {"504", 2, "Moroccan dirham", []string{"Morocco", "Western Sahara"}},
	"MDL": {"498", 2, "Moldovan leu", []string{"Moldova"}},
	"MGA": {"969", 2, "Malagasy ariary", []string{"Madagascar"}},
	"MKD": {"807", 2, "Macedonian denar", []string{"North Macedonia"}},
	"MMK": {"104", 2, "Myanmar kyat", []string{"Myanmar"}},
	"MNT": {"496", 2, "Mongolian tögrög", []string{"Mongolia"}},
	"MOP": {"446", 2, "Macanese pataca", []string{"Macau"}},
	"MRU": {"929", 2, "Mauritanian ouguiya", []string{"Mauritania"}},
	"MUR": {"480", 2, "Mauritian rupee", []string{"Mauritius"}},
	"MVR": {"462", 2, "Maldivian rufiyaa", []string{"Maldives"}},
	"MWK": {"454", 2, "Malawian kwacha", []string{"Malawi"}},
	"MXN": {"484", 2, "Mexican peso", []string{"Mexico"}},
	"MXV": {"979", 2, "Mexican Unidad de Inversion (UDI) (funds code)", []string{"Mexico"}},
	"MYR": {"458", 2, "Malaysian ringgit", []string{"Malaysia"}},
	"MZN": {"943", 2, "Mozambican metical", []string{"Mozambique"}},
	"NAD": {"516", 2, "Namibian dollar", []string{"Namibia (pegged to ZAR 1:1)"}},
	"NGN": {"566", 2, "Nigerian naira", []string{"Nigeria"}},
	"NIO": {"558", 2, "Nicaraguan córdoba", []string{"Nicaragua"}},
	"NOK": {"578", 2, "Norwegian krone", []string{"Norway", "Svalbard and  Jan Mayen (SJ)", "Bouvet Island (BV)"}},
	"NPR": {"524", 2, "Nepalese rupee", []string{"  Nepal"}},
	"NZD": {"554", 2, "New Zealand dollar", []string{"New Zealand", "Cook Islands (CK)", "Niue (NU)", "Pitcairn Islands (PN; see also Pitcairn Islands dollar)", "Tokelau (TK)"}},
	"OMR": {"512", 3, "Omani rial", []string{"Oman"}},
	"PAB": {"590", 2, "Panamanian balboa", []string{"Panama"}},
	"PEN": {"604", 2, "Peruvian sol", []string{"Peru"}},
	"PGK": {"598", 2, "Papua New Guinean kina", []string{"Papua New Guinea"}},
	"PHP": {"608", 2, "Philippine peso[11]", []string{"Philippines"}},
	"PKR": {"586", 2, "Pakistani rupee", []string{"Pakistan"}},
	"PLN": {"985", 2, "Polish złoty", []string{"Poland"}},
	"PYG": {"600", 0, "Paraguayan guaraní", []string{"Paraguay"}},
	"QAR": {"634", 2, "Qatari riyal", []string{"Qatar"}},
	"RON": {"946", 2, "Romanian leu", []string{"Romania"}},
	"RSD": {"941", 2, "Serbian dinar", []string{"Serbia"}},
	"RUB": {"643", 2, "Russian ruble", []string{"Russia"}},
	"RWF": {"646", 0, "Rwandan franc", []string{"Rwanda"}},
	"SAR": {"682", 2, "Saudi riyal", []string{"Saudi Arabia"}},
	"SBD": {"090", 2, "Solomon Islands dollar", []string{"Solomon Islands"}},
	"SCR": {"690", 2, "Seychelles rupee", []string{"Seychelles"}},
	"SDG": {"938", 2, "Sudanese pound", []string{"Sudan"}},
	"SEK": {"752", 2, "Swedish krona (plural: kronor)", []string{"Sweden"}},
	"SGD": {"702", 2, "Singapore dollar", []string{"Singapore"}},
	"SHP": {"654", 2, "Saint Helena pound", []string{"Saint Helena (SH-HL)", "Ascension Island (SH-AC)"}},
	"SLE": {"925", 2, "Sierra Leonean leone (new leone)[12][13][14]", []string{"Sierra Leone"}},
	"SOS": {"706", 2, "Somalian shilling", []string{"Somalia"}},
	"SRD": {"968", 2, "Surinamese dollar", []string{"Suriname"}},
	"SSP": {"728", 2, "South Sudanese pound", []string{"South Sudan"}},
	"STN": {"930", 2, "São Tomé and Príncipe dobra", []string{"São Tomé and Príncipe"}},
	"SVC": {"222", 2, "Salvadoran colón", []string{"El Salvador"}},
	"SYP": {"760", 2, "Syrian pound", []string{"Syria"}},
	"SZL": {"748", 2, "Swazi lilangeni", []string{"Eswatini[11]"}},
	"THB": {"764", 2, "Thai baht", []string{"Thailand"}},
	"TJS": {"972", 2, "Tajikistani somoni", []string{"Tajikistan"}},
	"TMT": {"934", 2, "Turkmenistan manat", []string{"Turkmenistan"}},
	"TND": {"788", 3, "Tunisian dinar", []string{"Tunisia"}},
	"TOP": {"776", 2, "Tongan paʻanga", []string{"Tonga"}},
	"TRY": {"949", 2, "Turkish lira", []string{"Turkey"}},
	"TTD": {"780", 2, "Trinidad and Tobago dollar", []string{"Trinidad and Tobago"}},
	"TWD": {"901", 2, "New Taiwan dollar", []string{"Taiwan"}},
	"TZS": {"834", 2, "Tanzanian shilling", []string{"Tanzania"}},
	"UAH": {"980", 2, "Ukrainian hryvnia", []string{"Ukraine"}},
	"UGX": {"800", 0, "Ugandan shilling", []string{"Uganda"}},
	"USD": {"840", 2, "United States dollar", []string{"United States", "American Samoa (AS)", "British Indian Ocean Territory (IO) (also uses GBP)", "British Virgin Islands (VG)", "Bonaire, Sint Eustatius and Saba (BQ - Caribbean Netherlands)", "Ecuador (EC)", "El Salvador (SV)", "Guam (GU)", "Marshall Islands (MH)", "Federated States of Micronesia (FM)", "Northern Mariana Islands (MP)", "Palau (PW)", "Panama (PA) (as well as Panamanian Balboa)", "Puerto Rico (PR)", "Timor-Leste (TL)", "Turks and Caicos Islands (TC)", "U.S. Virgin Islands (VI)", "United States Minor Outlying Islands (UM)"}},
	"USN": {"997", 2, "United States dollar (next day) (funds code)", []string{"United States"}},
	"UYI": {"940", 0, "Uruguay Peso en Unidades Indexadas (URUIURUI) (funds code)", []string{"Uruguay"}},
	"UYU": {"858", 2, "Uruguayan peso", []string{"Uruguay"}},
	"UYW": {"927", 4, "Unidad previsional[16]", []string{"Uruguay"}},
	"UZS": {"860", 2, "Uzbekistani sum", []string{"Uzbekistan"}},
	"VED": {"926", 2, "Venezuelan digital bolívar[17]", []string{"Venezuela"}},
	"VES": {"928", 2, "Venezuelan sovereign bolívar[11]", []string{"Venezuela"}},
	"VND": {"704", 0, "Vietnamese đồng", []string{"Vietnam"}},
	"VUV": {"548", 0, "Vanuatu vatu", []string{"Vanuatu"}},
	"WST": {"882", 2, "Samoan tala", []string{"Samoa"}},
	"XAF": {"950", 0, "CFA franc BEAC", []string{"Cameroon (CM)", "Central African Republic (CF)", "Republic of the Congo (CG)", "Chad (TD)", "Equatorial Guinea (GQ)", "Gabon (GA)"}},
	"XAG": {"961", 0, "Silver (one troy ounce)	", nil},
	"XAU": {"959", 0, "Gold (one troy ounce)	", nil},
	"XBA": {"955", 0, "European Composite Unit (EURCO) (bond market unit)	", nil},
	"XBB": {"956", 0, "European Monetary Unit (E.M.U.-6) (bond market unit)	", nil},
	"XBC": {"957", 0, "European Unit of Account 9 (E.U.A.-9) (bond market unit)	", nil},
	"XBD": {"958", 0, "European Unit of Account 17 (E.U.A.-17) (bond market unit)	", nil},
	"XCD": {"951", 2, "East Caribbean dollar", []string{"Anguilla (AI)", "Antigua and Barbuda (AG)", "Dominica (DM)", "Grenada (GD)", "Montserrat (MS)", "Saint Kitts and Nevis (KN)", "Saint Lucia (LC)", "Saint Vincent and the Grenadines (VC)"}},
	"XDR": {"960", 0, "Special drawing rights	International Monetary Fund", nil},
	"XOF": {"952", 0, "CFA franc BCEAO", []string{"Benin (BJ)", "Burkina Faso (BF)", "Côte d'Ivoire (CI)", "Guinea-Bissau (GW)", "Mali (ML)", "Niger (NE)", "Senegal (SN)", "Togo (TG)"}},
	"XPD": {"964", 0, "Palladium (one troy ounce)	", nil},
	"XPF": {"953", 0, "CFP franc (franc Pacifique)", []string{"French territories of the Pacific Ocean:  French Polynesia (PF)", "New Caledonia (NC)", "Wallis and Futuna (WF)"}},
	"XPT": {"962", 0, "Platinum (one troy ounce)	", nil},
	"XSU": {"994", 0, "SUCRE	Unified System for Regional Compensation (SUCRE)[18]", nil},
	"XTS": {"963", 0, "Code reserved for testing	", nil},
	"XUA": {"965", 0, "ADB Unit of Account	African Development Bank[19]", nil},
	"XXX": {"999", 0, "No currency	", nil},
	"YER": {"886", 2, "Yemeni rial", []string{"Yemen"}},
	"ZAR": {"710", 2, "South African rand", []string{"Eswatini", "Lesotho", "Namibia", "South Africa"}},
	"ZMW": {"967", 2, "Zambian kwacha", []string{"Zambia"}},
	"ZWG": {"924", 2, "Zimbabwe Gold", []string{"Zimbabwe[20]"}},
}
