"""Reference HTSUS subheadings Harbormaster ships with.

This is a curated slice of the tariff schedule covering the categories our design
partners import most (home goods, apparel, footwear, electronics, tools, auto parts,
furniture, metals). Rates are column-1 general (MFN) ad valorem rates or ad valorem
equivalents of specific rates, rounded. They are reference data for planning and
auditing; the production system syncs the full schedule from the USITC HTS release
and CBP CSMS messages, and every filing must be checked against the live HTSUS.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import date
from functools import cache

D = date


@dataclass(frozen=True)
class Step:
    since: date
    rate: float
    heading: str | None = None


@dataclass(frozen=True)
class HTSLine:
    code: str
    description: str
    mfn: float
    keywords: tuple[str, ...]
    # Section 232 coverage
    metal: str | None = None              # steel | aluminum | copper
    s232: str | None = None               # metal_primary | metal_derivative | auto | auto_part | truck
                                          # | lumber | upholstered | cabinet | semi | pharma
    derivative_since: date | None = None  # when a derivative entered the 232 annex
    s232_reduced: bool = False            # eligible for the 15% class from 2026-04-06
    # Section 301 (China) schedule for this line
    cn301: tuple[Step, ...] = ()
    # Why reciprocal / s122 surcharges did not apply, if they didn't
    recip_exempt: str | None = None       # annex_ii | electronics | ag_nov2025 | (None)
    recip_exempt_since: date | None = None
    adcvd_risk: str | None = None         # free-text note when AD/CVD orders commonly apply
    unit: str = "unit"
    tags: tuple[str, ...] = field(default=())

    @property
    def chapter(self) -> int:
        return int(self.code[:2])

    @property
    def heading(self) -> str:
        return self.code[:4]

    def cn301_rate(self, on: date) -> Step | None:
        current = None
        for step in self.cn301:
            if step.since <= on:
                current = step
        return current


L1 = (Step(D(2018, 7, 6), 25.0, "9903.88.01"),)
L2 = (Step(D(2018, 8, 23), 25.0, "9903.88.02"),)
L3 = (Step(D(2018, 9, 24), 10.0, "9903.88.03"), Step(D(2019, 5, 10), 25.0, "9903.88.03"))
L4A = (Step(D(2019, 9, 1), 15.0, "9903.88.15"), Step(D(2020, 2, 14), 7.5, "9903.88.15"))
NONE: tuple[Step, ...] = ()

AUG18 = D(2025, 8, 18)   # 407 derivative codes added to the steel/aluminum annex
MAR12 = D(2025, 3, 12)   # original derivative annex re-imposed for all countries
COPPER = D(2025, 8, 1)

LINES: tuple[HTSLine, ...] = (
    # ── Food & agriculture (reciprocal exemptions from 13 Nov 2025) ─────────────
    HTSLine("0901.21.00", "Coffee, roasted, not decaffeinated", 0.0,
            ("coffee", "roasted", "beans", "espresso"), recip_exempt="ag_nov2025",
            recip_exempt_since=D(2025, 11, 13), unit="kg"),
    HTSLine("0803.90.00", "Bananas, fresh or dried", 0.0, ("banana", "fruit", "fresh"),
            recip_exempt="ag_nov2025", recip_exempt_since=D(2025, 11, 13), unit="kg"),
    HTSLine("1806.32.90", "Chocolate in blocks or bars, not filled", 4.3,
            ("chocolate", "bar", "cocoa", "confectionery"), unit="kg"),
    HTSLine("2204.21.50", "Wine of fresh grapes, ≤ 2 litres, ≤ 14% alcohol", 2.6,
            ("wine", "red wine", "white wine", "bottle"), unit="litre"),
    HTSLine("0902.30.00", "Black tea, fermented, in packages ≤ 3 kg", 0.0,
            ("tea", "black tea", "tea bags"), recip_exempt="ag_nov2025",
            recip_exempt_since=D(2025, 11, 13), unit="kg"),
    # ── Pharma, cosmetics (Annex II / s232 pharma) ─────────────────────────────
    HTSLine("3004.90.92", "Medicaments, measured doses, other (incl. patented brands)", 0.0,
            ("medicine", "medicament", "tablets", "drug", "pharmaceutical", "capsules"),
            s232="pharma", recip_exempt="annex_ii", recip_exempt_since=D(2025, 4, 5), unit="kg"),
    HTSLine("3304.99.50", "Beauty, make-up and skin-care preparations, other", 0.0,
            ("skincare", "serum", "moisturizer", "cream", "cosmetic", "lotion", "beauty"),
            cn301=L4A, unit="kg"),
    HTSLine("3401.11.50", "Soap, toilet, in bars", 0.0, ("soap", "bar soap", "toilet soap"),
            cn301=L4A, unit="kg"),
    # ── Plastics & rubber ──────────────────────────────────────────────────────
    HTSLine("3923.21.00", "Sacks and bags of polymers of ethylene", 3.0,
            ("plastic bag", "poly bag", "mailer", "polyethylene bag", "sack"), cn301=L3),
    HTSLine("3924.10.40", "Tableware and kitchenware of plastics", 3.4,
            ("plastic kitchenware", "food container", "plastic tableware", "lunch box", "storage container"),
            cn301=L3),
    HTSLine("3926.90.99", "Other articles of plastics", 5.3,
            ("plastic", "plastic article", "phone case", "organizer", "plastic part"), cn301=L3),
    HTSLine("4015.12.10", "Medical, surgical, dental gloves of vulcanised rubber", 3.0,
            ("medical gloves", "nitrile gloves", "exam gloves", "surgical gloves"),
            cn301=(Step(D(2019, 9, 1), 15.0, "9903.88.15"), Step(D(2020, 2, 14), 7.5, "9903.88.15"),
                   Step(D(2025, 1, 1), 50.0, "9903.91.04"), Step(D(2026, 1, 1), 100.0, "9903.91.05")),
            adcvd_risk=None, unit="dozen pairs"),
    # ── Leather goods, luggage ─────────────────────────────────────────────────
    HTSLine("4202.12.81", "Suitcases and trunks, outer surface of man-made textile", 17.6,
            ("suitcase", "luggage", "carry-on", "trolley", "trunk"), cn301=L3),
    HTSLine("4202.21.90", "Handbags, outer surface of leather, > $20 each", 9.0,
            ("handbag", "leather bag", "purse", "tote leather"), cn301=L3),
    HTSLine("4202.92.31", "Travel, sports and similar bags, outer surface of man-made textile", 17.6,
            ("backpack", "duffel", "gym bag", "travel bag", "tote"), cn301=L3),
    # ── Wood, lumber, paper ────────────────────────────────────────────────────
    HTSLine("4407.11.00", "Lumber of pine, sawn lengthwise, > 6 mm thick", 0.0,
            ("lumber", "pine", "softwood", "timber", "boards"), s232="lumber",
            recip_exempt="annex_ii", recip_exempt_since=D(2025, 4, 5),
            adcvd_risk="Softwood lumber from Canada is subject to AD/CVD orders.", unit="m3"),
    HTSLine("4421.99.98", "Other articles of wood", 3.3,
            ("wood", "wooden", "cutting board", "wood decor", "wood tray"), cn301=L3),
    HTSLine("4819.10.00", "Cartons, boxes and cases of corrugated paper", 0.0,
            ("box", "carton", "corrugated", "shipping box", "packaging"), cn301=L3),
    # ── Apparel ────────────────────────────────────────────────────────────────
    HTSLine("6109.10.00", "T-shirts, singlets and tank tops, knitted, of cotton", 16.5,
            ("t-shirt", "tee", "tank top", "cotton shirt", "knit top"), cn301=L4A),
    HTSLine("6110.20.20", "Sweaters, pullovers, sweatshirts, knitted, of cotton", 16.5,
            ("sweater", "pullover", "sweatshirt", "hoodie", "jumper", "cardigan"), cn301=L4A),
    HTSLine("6104.43.20", "Women's or girls' dresses, knitted, of synthetic fibres", 16.0,
            ("dress", "women's dress", "knit dress", "jersey dress"), cn301=L4A),
    HTSLine("6203.42.40", "Men's trousers and shorts, of cotton, not knitted", 16.6,
            ("jeans", "trousers", "pants", "chinos", "denim", "shorts"), cn301=L4A),
    HTSLine("6204.62.40", "Women's trousers and shorts, of cotton, not knitted", 16.6,
            ("women's jeans", "women's trousers", "women's pants", "women's shorts"), cn301=L4A),
    HTSLine("6201.40.10", "Men's anoraks and jackets, of man-made fibres, not knitted", 27.7,
            ("jacket", "anorak", "parka", "puffer", "windbreaker", "coat"), cn301=L4A),
    HTSLine("6115.95.90", "Socks and hosiery, knitted, of cotton", 13.5,
            ("socks", "hosiery", "ankle socks"), cn301=L4A, unit="dozen pairs"),
    HTSLine("6302.31.90", "Bed linen of cotton, not printed", 6.7,
            ("bed sheet", "sheets", "duvet cover", "pillowcase", "bedding", "linen"), cn301=L4A),
    # ── Footwear ───────────────────────────────────────────────────────────────
    HTSLine("6404.11.90", "Sports footwear, textile uppers, rubber/plastic soles, > $12/pair", 20.0,
            ("sneakers", "running shoes", "trainers", "athletic shoes", "shoes"), cn301=L4A,
            unit="pairs"),
    HTSLine("6403.99.90", "Footwear with leather uppers, other, > $2.50/pair", 10.0,
            ("leather shoes", "boots", "loafers", "dress shoes"), cn301=L4A, unit="pairs"),
    HTSLine("6402.99.31", "Footwear, rubber or plastic uppers, other (sandals, slides)", 6.0,
            ("sandals", "slides", "flip flops", "clogs", "rubber shoes"), cn301=L4A, unit="pairs"),
    # ── Ceramics, glass, jewelry ───────────────────────────────────────────────
    HTSLine("6912.00.48", "Ceramic tableware and kitchenware, other than porcelain", 9.8,
            ("mug", "ceramic", "stoneware", "plate", "bowl", "dinnerware"), cn301=L3),
    HTSLine("6911.10.80", "Porcelain or china tableware, other", 20.8,
            ("porcelain", "china plate", "fine china", "teacup"), cn301=L3),
    HTSLine("7013.37.20", "Drinking glasses, other than lead crystal", 7.5,
            ("glass", "drinking glass", "tumbler", "glassware", "wine glass"), cn301=L3),
    HTSLine("7113.19.50", "Jewellery of precious metal other than silver", 5.5,
            ("jewelry", "gold ring", "necklace", "earrings", "bracelet"), cn301=L4A),
    # ── Iron & steel ───────────────────────────────────────────────────────────
    HTSLine("7208.39.00", "Flat-rolled iron/steel, hot-rolled, in coils, < 3 mm", 0.0,
            ("hot-rolled steel", "steel coil", "steel sheet", "flat steel"),
            metal="steel", s232="metal_primary",
            cn301=(Step(D(2018, 9, 24), 10.0, "9903.88.03"), Step(D(2019, 5, 10), 25.0, "9903.88.03")),
            adcvd_risk="Hot-rolled steel is subject to AD/CVD orders from several countries.", unit="kg"),
    HTSLine("7306.30.50", "Welded tubes and pipes of circular cross-section, iron or steel", 0.0,
            ("steel pipe", "tube", "welded pipe"), metal="steel", s232="metal_primary", cn301=L3,
            adcvd_risk="Circular welded pipe is subject to AD/CVD orders.", unit="kg"),
    HTSLine("7318.15.20", "Bolts of iron or steel, with nuts or washers", 0.0,
            ("bolt", "fastener", "screw", "nut", "hardware"), metal="steel",
            s232="metal_derivative", derivative_since=MAR12, cn301=L3, unit="kg"),
    HTSLine("7323.93.00", "Table, kitchen or household articles of stainless steel", 2.0,
            ("stainless steel", "cookware", "pot", "pan", "water bottle", "tumbler steel", "kitchenware steel"),
            metal="steel", s232="metal_derivative", derivative_since=AUG18, cn301=L3),
    HTSLine("7326.90.86", "Other articles of iron or steel", 2.9,
            ("steel article", "metal bracket", "steel rack", "hook", "steel shelf"), metal="steel",
            s232="metal_derivative", derivative_since=AUG18, cn301=L3),
    HTSLine("7321.11.10", "Cooking appliances for gas fuel, portable (grills, stoves)", 5.7,
            ("grill", "barbecue", "camping stove", "gas stove", "bbq"), metal="steel",
            s232="metal_derivative", derivative_since=AUG18, cn301=L3),
    # ── Copper ─────────────────────────────────────────────────────────────────
    HTSLine("7408.11.60", "Refined copper wire, max cross-section > 6 mm", 3.0,
            ("copper wire", "copper rod", "copper"), metal="copper", s232="metal_primary",
            cn301=L3, recip_exempt="annex_ii", recip_exempt_since=D(2025, 4, 5), unit="kg"),
    HTSLine("7411.10.10", "Tubes and pipes of refined copper", 1.5,
            ("copper pipe", "copper tube", "refrigeration tube"), metal="copper",
            s232="metal_primary", cn301=L3, recip_exempt="annex_ii",
            recip_exempt_since=D(2025, 4, 5), unit="kg"),
    # ── Aluminum ───────────────────────────────────────────────────────────────
    HTSLine("7604.29.10", "Aluminium alloy bars, rods and profiles (extrusions)", 5.0,
            ("aluminum extrusion", "aluminium profile", "aluminum bar", "aluminum rod"),
            metal="aluminum", s232="metal_primary", cn301=L3,
            adcvd_risk="Aluminum extrusions from China are subject to AD/CVD orders.", unit="kg"),
    HTSLine("7615.10.71", "Cooking and kitchen ware of cast aluminium", 3.1,
            ("aluminum pan", "cast aluminum", "aluminum cookware", "dutch oven"), metal="aluminum",
            s232="metal_derivative", derivative_since=AUG18, cn301=L3),
    HTSLine("7616.99.51", "Other articles of aluminium", 2.5,
            ("aluminum", "aluminium article", "aluminum frame", "aluminum case"), metal="aluminum",
            s232="metal_derivative", derivative_since=MAR12, cn301=L3),
    # ── Tools & hardware ───────────────────────────────────────────────────────
    HTSLine("8205.59.80", "Hand tools, other, of base metal", 5.3,
            ("hand tool", "wrench", "pliers", "screwdriver", "tool set"), metal="steel",
            s232="metal_derivative", derivative_since=AUG18, cn301=L3),
    HTSLine("8467.21.00", "Drills of all kinds, electromechanical, self-contained motor", 1.7,
            ("drill", "cordless drill", "power tool", "impact driver"), cn301=L3),
    HTSLine("8302.42.30", "Base metal mountings and fittings suitable for furniture", 3.9,
            ("hinge", "drawer slide", "furniture fitting", "handle", "knob"), metal="steel",
            s232="metal_derivative", derivative_since=AUG18, cn301=L3),
    # ── Machinery & appliances ─────────────────────────────────────────────────
    HTSLine("8414.51.30", "Table, floor, wall, ceiling fans, ≤ 125 W", 4.7,
            ("fan", "desk fan", "ceiling fan", "tower fan"), cn301=L3),
    HTSLine("8415.10.30", "Window or wall air conditioners, self-contained", 2.2,
            ("air conditioner", "window ac", "portable ac", "hvac"), s232="metal_derivative",
            metal="steel", derivative_since=AUG18, s232_reduced=True, cn301=L3),
    HTSLine("8418.21.00", "Household refrigerators, compression type", 1.0,
            ("refrigerator", "fridge", "mini fridge"), metal="steel", s232="metal_derivative",
            derivative_since=AUG18, cn301=L3),
    HTSLine("8413.70.20", "Centrifugal pumps, single-stage, single-suction", 0.0,
            ("pump", "water pump", "centrifugal pump"), metal="steel", s232="metal_derivative",
            derivative_since=AUG18, s232_reduced=True, cn301=L1),
    HTSLine("8481.80.10", "Hand-operated valves of copper or copper alloy", 4.0,
            ("valve", "brass valve", "faucet", "tap", "ball valve"), metal="copper",
            s232="metal_derivative", derivative_since=COPPER, cn301=L2),
    HTSLine("8482.10.50", "Ball bearings, other", 9.0,
            ("bearing", "ball bearing", "roller bearing"), metal="steel", s232="metal_derivative",
            derivative_since=AUG18, cn301=L1,
            adcvd_risk="Ball bearings from some origins have had AD orders."),
    HTSLine("8501.52.40", "AC motors, multi-phase, > 750 W but ≤ 75 kW", 2.8,
            ("electric motor", "ac motor", "induction motor"), metal="copper",
            s232="metal_derivative", derivative_since=COPPER, s232_reduced=True, cn301=L1),
    HTSLine("8509.40.00", "Food grinders, processors, mixers and blenders, domestic", 4.2,
            ("blender", "mixer", "food processor", "juicer"), cn301=L3),
    HTSLine("8516.71.00", "Coffee or tea makers, electrothermic, domestic", 2.5,
            ("coffee maker", "espresso machine", "kettle", "tea maker"), cn301=L3),
    HTSLine("8516.31.00", "Electrothermic hair dryers", 3.9,
            ("hair dryer", "blow dryer", "hair styler"), cn301=L4A),
    HTSLine("8433.11.00", "Lawn mowers, powered, cutting device rotating horizontally", 0.0,
            ("lawn mower", "mower", "robot mower"), metal="steel", s232="metal_derivative",
            derivative_since=AUG18, s232_reduced=True, cn301=L3),
    HTSLine("8443.32.10", "Printers capable of connecting to a computer", 0.0,
            ("printer", "label printer", "3d printer", "inkjet"), cn301=L4A),
    HTSLine("8471.30.01", "Portable automatic data-processing machines ≤ 10 kg (laptops)", 0.0,
            ("laptop", "notebook computer", "chromebook", "tablet computer"),
            recip_exempt="electronics", recip_exempt_since=D(2025, 4, 5)),
    HTSLine("8473.30.11", "Parts and accessories of computers, printed circuit assemblies", 0.0,
            ("computer parts", "motherboard", "pcb assembly", "graphics card"), cn301=L3,
            recip_exempt="electronics", recip_exempt_since=D(2025, 4, 5)),
    # ── Electrical & electronics ───────────────────────────────────────────────
    HTSLine("8504.40.95", "Static converters (power supplies, chargers), other", 0.0,
            ("charger", "power adapter", "power supply", "usb charger", "inverter"), cn301=L3),
    HTSLine("8507.60.00", "Lithium-ion accumulators (non-EV)", 3.4,
            ("lithium battery", "battery pack", "power bank", "li-ion", "rechargeable battery"),
            cn301=(Step(D(2019, 9, 1), 15.0, "9903.88.15"), Step(D(2020, 2, 14), 7.5, "9903.88.15"),
                   Step(D(2026, 1, 1), 25.0, "9903.91.06"))),
    HTSLine("8517.13.00", "Smartphones", 0.0, ("smartphone", "mobile phone", "cell phone", "iphone", "android phone"),
            recip_exempt="electronics", recip_exempt_since=D(2025, 4, 5)),
    HTSLine("8518.30.20", "Headphones and earphones", 0.0,
            ("headphones", "earbuds", "earphones", "headset"), cn301=L4A),
    HTSLine("8518.22.00", "Multiple loudspeakers, mounted in the same enclosure", 4.9,
            ("speaker", "bluetooth speaker", "soundbar", "loudspeaker"), cn301=L3),
    HTSLine("8528.72.64", "Colour television receivers, flat panel", 3.9,
            ("tv", "television", "monitor tv", "smart tv"), cn301=L4A),
    HTSLine("8539.52.00", "LED lamps", 2.0, ("led bulb", "light bulb", "led lamp"), cn301=L4A),
    HTSLine("8541.43.00", "Photovoltaic cells assembled in modules (solar panels)", 0.0,
            ("solar panel", "solar module", "photovoltaic", "pv module"),
            cn301=(Step(D(2018, 9, 24), 10.0, "9903.88.03"), Step(D(2019, 5, 10), 25.0, "9903.88.03"),
                   Step(D(2024, 9, 27), 50.0, "9903.91.01")),
            recip_exempt="annex_ii", recip_exempt_since=D(2025, 4, 5),
            adcvd_risk="Solar cells/modules from several Asian origins are subject to AD/CVD and s201."),
    HTSLine("8542.31.00", "Processors and controllers (integrated circuits)", 0.0,
            ("processor", "chip", "microcontroller", "integrated circuit", "gpu chip", "semiconductor"),
            s232="semi",
            cn301=(Step(D(2018, 8, 23), 25.0, "9903.88.02"), Step(D(2025, 1, 1), 50.0, "9903.91.02")),
            recip_exempt="annex_ii", recip_exempt_since=D(2025, 4, 5)),
    HTSLine("8544.42.90", "Insulated electric conductors with connectors, ≤ 1,000 V", 2.6,
            ("cable", "usb cable", "power cord", "hdmi cable", "wire harness"), metal="copper",
            s232="metal_derivative", derivative_since=COPPER, cn301=L3),
    # ── Vehicles & parts ───────────────────────────────────────────────────────
    HTSLine("8703.80.00", "Passenger vehicles with only electric motor", 2.5,
            ("electric vehicle", "ev", "electric car"), s232="auto",
            cn301=(Step(D(2018, 7, 6), 25.0, "9903.88.01"), Step(D(2024, 9, 27), 100.0, "9903.91.20"))),
    HTSLine("8703.23.01", "Passenger vehicles, spark-ignition, 1,500–3,000 cc", 2.5,
            ("car", "sedan", "suv", "passenger vehicle"), s232="auto", cn301=L1),
    HTSLine("8704.23.01", "Trucks for goods, diesel, GVW > 20 t", 25.0,
            ("heavy truck", "tractor truck", "semi truck", "lorry"), s232="truck", cn301=L1),
    HTSLine("8708.29.50", "Parts and accessories of motor-vehicle bodies, other", 2.5,
            ("auto body part", "bumper", "door panel", "car part", "fender"), s232="auto_part", cn301=L3),
    HTSLine("8708.30.50", "Brakes and servo-brakes and parts, other", 2.5,
            ("brake", "brake pad", "rotor", "caliper"), s232="auto_part", cn301=L1),
    HTSLine("8711.60.00", "Motorcycles and cycles with electric motor (e-bikes)", 0.0,
            ("e-bike", "electric bike", "electric scooter", "electric motorcycle"), cn301=L3),
    HTSLine("8712.00.15", "Bicycles, not motorised, wheels ≤ 63.5 cm", 5.5,
            ("bicycle", "bike", "kids bike"), cn301=L3),
    # ── Instruments ────────────────────────────────────────────────────────────
    HTSLine("9004.10.00", "Sunglasses", 2.0, ("sunglasses", "shades", "eyewear"), cn301=L4A),
    HTSLine("9018.31.00", "Syringes, with or without needles", 0.0,
            ("syringe", "needle", "injection"),
            cn301=(Step(D(2018, 9, 24), 10.0, "9903.88.03"), Step(D(2019, 5, 10), 25.0, "9903.88.03"),
                   Step(D(2024, 9, 27), 100.0, "9903.91.03"))),
    HTSLine("9018.90.80", "Medical, surgical instruments and appliances, other", 0.0,
            ("medical device", "surgical instrument", "medical instrument"), cn301=L3),
    # ── Furniture & lighting ───────────────────────────────────────────────────
    HTSLine("9401.61.40", "Upholstered seats with wooden frames, other", 0.0,
            ("sofa", "couch", "armchair", "upholstered chair", "loveseat", "sectional"),
            s232="upholstered", cn301=L3),
    HTSLine("9401.71.00", "Upholstered seats with metal frames", 0.0,
            ("office chair", "metal chair upholstered", "gaming chair"), cn301=L3,
            metal="steel", s232="metal_derivative", derivative_since=AUG18),
    HTSLine("9403.40.90", "Wooden furniture of a kind used in the kitchen (cabinets)", 0.0,
            ("kitchen cabinet", "cabinet", "vanity", "bathroom vanity"), s232="cabinet", cn301=L3,
            adcvd_risk="Wooden cabinets and vanities from China are subject to AD/CVD orders."),
    HTSLine("9403.60.80", "Wooden furniture, other", 0.0,
            ("table", "desk", "bookshelf", "dresser", "wood furniture", "nightstand", "bed frame"),
            cn301=L3),
    HTSLine("9403.20.00", "Metal furniture, other", 0.0,
            ("metal furniture", "metal shelf", "steel desk", "metal bed"), metal="steel",
            s232="metal_derivative", derivative_since=AUG18, cn301=L3),
    HTSLine("9404.21.00", "Mattresses of cellular rubber or plastics", 3.0,
            ("mattress", "memory foam", "foam mattress"), cn301=L3,
            adcvd_risk="Mattresses from several origins are subject to AD/CVD orders."),
    HTSLine("9405.11.40", "Chandeliers and ceiling lights, LED, of base metal", 3.9,
            ("chandelier", "ceiling light", "pendant light", "light fixture"), cn301=L3),
    HTSLine("9405.21.80", "Electric table, desk, bedside or floor lamps, LED", 3.9,
            ("lamp", "desk lamp", "floor lamp", "table lamp"), cn301=L3),
    # ── Toys, games, sports ────────────────────────────────────────────────────
    HTSLine("9503.00.00", "Toys (tricycles, dolls, puzzles, scale models, other)", 0.0,
            ("toy", "doll", "puzzle", "plush", "lego", "building blocks", "game for kids")),
    HTSLine("9504.50.00", "Video game consoles and machines", 0.0,
            ("game console", "video game", "handheld console")),
    HTSLine("9506.91.00", "Articles and equipment for general physical exercise", 4.6,
            ("exercise equipment", "dumbbell", "yoga mat", "treadmill", "resistance band", "kettlebell"),
            cn301=L4A),
    HTSLine("9506.62.80", "Inflatable balls, other", 0.0, ("ball", "soccer ball", "basketball", "volleyball"),
            cn301=L4A),
    HTSLine("9603.29.40", "Hair, shaving and nail brushes", 0.0,
            ("brush", "hair brush", "makeup brush", "shaving brush"), cn301=L4A),
)


@cache
def by_code() -> dict[str, HTSLine]:
    return {line.code: line for line in LINES}


def lookup(code: str) -> HTSLine | None:
    """Find a reference line by 8- or 10-digit code, with or without dots."""
    digits = "".join(ch for ch in (code or "") if ch.isdigit())
    if len(digits) < 8:
        return None
    key = f"{digits[:4]}.{digits[4:6]}.{digits[6:8]}"
    return by_code().get(key)


def search(query: str, limit: int = 8) -> list[tuple[HTSLine, float]]:
    """Keyword search over the reference catalog. Returns (line, score) best first.

    This is the deterministic classifier used when no model is configured and as a
    candidate generator for the AI classifier."""
    q = (query or "").lower()
    words = {w.strip(".,;:()/-") for w in q.split() if len(w) > 2}
    scored: list[tuple[HTSLine, float]] = []
    for line in LINES:
        score = 0.0
        for kw in line.keywords:
            if kw in q:
                score += 3.0 + 0.5 * kw.count(" ")
        desc = line.description.lower()
        score += sum(0.6 for w in words if w in desc)
        if score > 0:
            scored.append((line, score))
    scored.sort(key=lambda t: -t[1])
    return scored[:limit]
