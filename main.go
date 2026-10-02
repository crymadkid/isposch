package main

import (
	"fmt"
	"isposch/parser"
)

func main() {
	var testXML = []byte(`
		<dataroot">
			<My>
				<ID>1278</ID>
				<DAT>2026-09-10T00:00:00</DAT>
				<UR>3</UR>
				<IDGG>2</IDGG>
				<FAMIO>Большакова А.В.</FAMIO>
				<SPPRED.NAIM>Ин. яз.</SPPRED.NAIM>
				<SPGRUP.NAIM>26294315/2092</SPGRUP.NAIM>
				<ZAM>2</ZAM>
				<KONS>0</KONS>
				<AUD/>
				<NOTE/>
				<CAMPUS>П</CAMPUS>
			</My>
		</dataroot>
	`)

	lessons, err := parser.ParseSchedule(testXML)
	if err != nil {
		fmt.Println("Ошибка парсинга:", err)
		return
	}

	fmt.Printf("Распарсено записей: %d\n", len(lessons))

	if len(lessons) > 0 {
		fmt.Printf("Первый предмет: %s, Преподаватель: %s\n",
			lessons[0].Subject,
			lessons[0].Teacher,
		)
	}
}
