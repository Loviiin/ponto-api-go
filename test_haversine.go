package main

import (
	"fmt"

	"github.com/umahmood/haversine"
)

func main() {
	// Coordenadas da empresa (DB)
	empresaCoord := haversine.Coord{Lat: -15.8048986, Lon: -47.8756844}

	// Coordenadas da batida de ponto (novas fornecidas)
	batidaCoord := haversine.Coord{Lat: -15.7990468, Lon: -47.8802718}

	// Calcular distância
	km, _ := haversine.Distance(empresaCoord, batidaCoord)
	distanciaEmMetros := km * 1000

	fmt.Printf("Coordenadas da Empresa: lat=%.12f, lon=%.12f\n", empresaCoord.Lat, empresaCoord.Lon)
	fmt.Printf("Coordenadas da Batida:  lat=%.12f, lon=%.12f\n", batidaCoord.Lat, batidaCoord.Lon)
	fmt.Printf("Distância calculada: %.2f km (%.2f metros)\n", km, distanciaEmMetros)
	fmt.Printf("Raio geofence configurado: 150 metros\n")

	raioGeofence := 150
	if distanciaEmMetros > float64(raioGeofence) {
		fmt.Printf("Resultado: REMOTO (distância %.2f m > raio %.2f m)\n", distanciaEmMetros, float64(raioGeofence))
	} else {
		fmt.Printf("Resultado: PRESENCIAL (distância %.2f m <= raio %.2f m)\n", distanciaEmMetros, float64(raioGeofence))
	}
}
