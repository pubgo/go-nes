package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/k0kubun/pp"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"os"
)

func main() {
	filename := os.Args[1]
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(b))
	fmt.Println(string(b[0]) + string(b[1]) + string(b[2]))
	prgSize := int(b[4])
	chrSize := int(b[5])
	prgRomEnd := 0x10 + prgSize*0x4000
	prgRom := b[0x10:prgRomEnd]
	chrRom := b[prgRomEnd : prgRomEnd+chrSize*0x2000]
	fmt.Printf("PRG SIZE: %d => %d\n", prgSize, len(prgRom))
	fmt.Printf("CHR SIZE: %d => %d\n", chrSize, len(chrRom))

	cpu := &Cpu{
		RAM: make([]int, 0x0800),
		Register: &Register{
			P: &StatusRegister{},
		},
		PrgROM: prgRom,
	}

	ppu := NewPPU()
	sprites := make([]*Sprite, 512)
	for i := 0; i < 512; i++ {
		index := i * 16
		sprites[i] = NewSprite(chrRom[index : index+16])
	}
	ppu.sprites = sprites
	cpu.PPU = ppu
	cpu.Reset()
	nes := &NES{
		cpu: cpu,
		ppu: ppu,
	}
	if err := ebiten.Run(nes.update, 256, 240, 2, "sample"); err != nil {
		log.Fatal(err)
	}
}

type NES struct {
	cpu        *Cpu
	ppu        *PPU
	background *BackGround
	pallet     *Pallet
}

func (nes *NES) update(screen *ebiten.Image) error {
	if ebiten.IsDrawingSkipped() {
		return nil
	}

	for {
		cycle := nes.cpu.Run()
		background, pallet := nes.ppu.Run(cycle * 3)
		if background != nil {
			nes.background = background
			nes.pallet = pallet
		}
		if nes.background != nil {
			nes.renderEbiten(screen, nes.background, nes.pallet)
			break
		}
	}
	return nil
}

func (nes *NES) renderEbiten(screen *ebiten.Image, background *BackGround, pallet *Pallet) {
	for i, line := range background.tiles {
		for j, tile := range line {
			for y, line := range tile.img.bitMap {
				for x, bit := range line {
					if bit != 0 {
						img, _ := ebiten.NewImage(1, 1, 0)
						c := pallet.getColor(tile.palletId, bit)
						img.Fill(color.RGBA{c.R, c.G, c.B, 0xff})
						options := &ebiten.DrawImageOptions{}
						options.GeoM.Translate(float64(j*SpriteSize+x), float64(i*SpriteSize+y))
						screen.DrawImage(img, options)
					}
				}
			}
		}
	}
}

type Register struct {
	A  int
	X  int
	Y  int
	P  *StatusRegister
	SP int
	PC int
}

type BackGround struct {
	tiles [][]*Tile
}

func (b *BackGround) Add(x, y int, tile *Tile) {
	b.tiles[y][x] = tile
}

func NewBackGround() *BackGround {
	tiles := make([][]*Tile, 30)
	for i := 0; i < 30; i++ {
		tiles[i] = make([]*Tile, 32)
	}
	return &BackGround{
		tiles: tiles,
	}
}

type Tile struct {
	img      *Sprite
	palletId int
}

type StatusRegister struct {
	Negative  bool
	Overflow  bool
	Reserved  bool
	Break     bool
	Decimal   bool
	Interrupt bool
	Zero      bool
	Carry     bool
}

func (r *StatusRegister) Int() int {
	return bool2int(r.Negative)*int(math.Pow(2, 7)) +
		bool2int(r.Overflow)*int(math.Pow(2, 6)) +
		bool2int(r.Reserved)*int(math.Pow(2, 5)) +
		bool2int(r.Break)*int(math.Pow(2, 4)) +
		bool2int(r.Decimal)*int(math.Pow(2, 3)) +
		bool2int(r.Interrupt)*int(math.Pow(2, 2)) +
		bool2int(r.Zero)*int(math.Pow(2, 1)) +
		bool2int(r.Carry)*int(math.Pow(2, 0))
}

func (r *StatusRegister) Set(v int) {
	r.Negative = int(math.Pow(2, 7)) != 0
	r.Overflow = int(math.Pow(2, 6)) != 0
	r.Reserved = int(math.Pow(2, 5)) != 0
	r.Break = int(math.Pow(2, 4)) != 0
	r.Decimal = int(math.Pow(2, 3)) != 0
	r.Interrupt = int(math.Pow(2, 2)) != 0
	r.Zero = int(math.Pow(2, 1)) != 0
	r.Carry = int(math.Pow(2, 0)) != 0
}

func debug(args ...interface{}) {
	if true {
		pp.Println(args...)
	}
}

func bool2int(v bool) int {
	if v {
		return 1
	}
	return 0
}
