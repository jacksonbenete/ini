package ini

import (
	"fmt"
	"slices"
	"testing"
)

func TestSimpleEncoding(t *testing.T) {
	type Object struct {
		Name string `ini:"name"`
	}

	data := `name=foo`

	o := Object{}

	err := Unmarshal([]byte(data), &o)
	if err != nil {
		t.Fatalf("failed decoding %v", err)
	}

	if o.Name != "foo" {
		t.Fatalf("expected %s, got %s", "foo", o.Name)
	}
}

func TestTaggedEncoding(t *testing.T) {
	type Profession struct {
		A string `ini:"a"`
		B bool   `ini:"b"`
	}

	type Object struct {
		Name       string     `ini:"name"`
		Profession Profession `ini:"profession"`
	}

	data := `
name=foo

[profession]
a=some string
b=true`

	o := Object{}

	err := Unmarshal([]byte(data), &o)
	if err != nil {
		t.Fatalf("failed decoding %v", err)
	}

	if o.Name != "foo" {
		t.Fatalf("expected %s, got %s", "foo", o.Name)
	}

	if o.Profession.A != "some string" {
		t.Fatalf("expected %s, got %s", "some string", o.Profession.A)

	}

	if !o.Profession.B {
		t.Fatalf("expected %t, got %+v", true, o.Profession.B)

	}
}

func TestWithDefault(t *testing.T) {
	type A struct {
		Value int `ini:"value"`
	}
	type B struct {
		Value bool `ini:"value"`
	}
	type C struct {
		Value float64 `ini:"value"`
	}
	type D struct {
		Value int `ini:"value"`
	}
	type Object struct {
		A A `ini:"a"`
		B B `ini:"b"`
		C C `ini:"c"`
		D D `ini:"d"`
	}

	data := `
[a]
value=3

[b]
value=true`

	o := Object{A: A{Value: 1}, C: C{Value: 1.0}}

	err := Unmarshal([]byte(data), &o)
	if err != nil {
		t.Fatalf("failed decoding %v", err)
	}

	if o.D.Value != 0 {
		t.Fatal("expected default value 0, got", o.D.Value)
	}

	if o.C.Value != 1.0 {
		t.Fatal("expected pre-set value 1.0, got ", o.C.Value)
	}

	if o.B.Value != true {
		t.Fatal("expected assigned value true, got ", o.B.Value)
	}

	if o.A.Value != 3 {
		t.Fatal("expected overwrite value 3, got", o.A.Value)
	}
}

type Config struct {
	DataDir  string
	Window   WindowConfig   `ini:"window"`
	Packet   PacketConfig   `ini:"packet"`
	Login    LoginConfig    `ini:"login"`
	Audio    AudioConfig    `ini:"audio"`
	Render   RenderConfig   `ini:"render"`
	Network  NetworkConfig  `ini:"network"`
	Fog      FogConfig      `ini:"fog"`
	Gameplay GameplayConfig `ini:"gameplay"`
	Script   ScriptConfig   `ini:"script"`
}

type WindowConfig struct {
	Title      string `ini:"title"`
	Width      int    `ini:"width"`
	Height     int    `ini:"height"`
	Fullscreen bool   `ini:"fullscreen"`
}

type PacketConfig struct {
	ClientDate int `ini:"client_date"`
	Profile    int `ini:"profile"`
}

type LoginConfig struct {
	Username  string `ini:"username"`
	Password  string `ini:"password"`
	AutoLogin bool   `ini:"auto_login"`
	CharSlot  int
}

type AudioConfig struct {
	Disabled  bool    `ini:"disable"`
	BGM       bool    `ini:"bgm"`
	BGMVolume float64 `ini:"bgm_volume"`
	SFXVolume float64 `ini:"sfx_volume"`
}

type RenderConfig struct {
	GraphicsAPI        string `ini:"graphics_api"`
	VSync              bool   `ini:"vsync"`
	FPS                bool   `ini:"fps"`
	NoUI               bool   `ini:"no_ui"`
	AsyncUI            bool   `ini:"async_ui"`
	UIProfile          bool   `ini:"ui_profile"`
	BenchSeconds       int    `ini:"bench_seconds"`
	BenchWarmupSeconds int    `ini:"bench_warmuop_seconds"`
	CPUProfile         string `ini:"cpu_profile"`
	Stats              bool   `ini:"stats"`
	WorldDebugStats    bool   `ini:"world_debug_stats"`
}

type NetworkConfig struct {
	Trace bool `ini:"trace"`
}

type FogConfig struct {
	Enabled bool `ini:"enabled"`
}

type GameplayConfig struct {
	NoShift     bool `ini:"no_shift"`
	NoCtrl      bool `ini:"no_ctrl"`
	LessEffects bool `ini:"less_effects"`
	SnapTargets bool `ini:"snap_targets"`
	SnapItems   bool `ini:"snap_items"`
	ForceUserAI bool `ini:"force_user_ai"`
}

type ScriptConfig struct {
	Path string
}

func TestComplexStructure(t *testing.T) {
	expectedAutoLogin := true
	expectedBgmVolume := 0.55
	expectedVsync := true

	goro_ini := fmt.Sprintf(`
[login]
username=a
password=
auto_login=%v

[window]
fullscreen = false

[render]
vsync = %v
fps = false

[audio]
bgm_volume = %.2f
sfx_volume = 0.55

[gameplay]
no_shift = false
no_ctrl = true
less_effects = false
snap = false
itemsnap = false`, expectedAutoLogin, expectedVsync, expectedBgmVolume)

	cfg := Config{
		Window: WindowConfig{
			Title:      "goro",
			Width:      1280,
			Height:     720,
			Fullscreen: false,
		},
		Packet: PacketConfig{
			ClientDate: 20080910,
			Profile:    23,
		},
		Login: LoginConfig{
			CharSlot: -1,
		},
		Audio: AudioConfig{
			BGM:       true,
			BGMVolume: 0.55,
			SFXVolume: 0.55,
		},
		Render: RenderConfig{
			GraphicsAPI:        "vulkan",
			AsyncUI:            true,
			VSync:              true,
			BenchWarmupSeconds: 0,
		},
		Fog: FogConfig{
			Enabled: true,
		},
		Gameplay: GameplayConfig{
			NoCtrl: true,
		},
	}

	err := Unmarshal([]byte(goro_ini), &cfg)
	if err != nil {
		t.Fatalf("cannot unmarshal ini: %v", err)
	}

	if cfg.Window.Width <= 0 {
		t.Fatalf("expected unmarshal to preserve struct values, expect Width: 1280, got %d", cfg.Window.Width)
	}
	if cfg.Login.AutoLogin != expectedAutoLogin {
		t.Fatalf("expected %v, got %v", expectedAutoLogin, cfg.Login.AutoLogin)
	}
	if cfg.Audio.BGMVolume != 0.55 {
		t.Fatalf("expected %v, got %v", expectedBgmVolume, cfg.Audio.BGMVolume)
	}
	if cfg.Render.VSync != true {
		t.Fatalf("expected %v, got %v", expectedVsync, cfg.Render.VSync)
	}
}

type UIConfig struct {
	WindowDragMode string `ini:"window_drag_mode"`
}

type SomeConfig struct {
	UI UIConfig `ini:"ui"`
}

func TestConfigWithSpace(t *testing.T) {
	file := []byte(`[ui]
window_drag_mode = drop
`)

	var cfg SomeConfig
	if err := Unmarshal(file, &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.UI.WindowDragMode != "drop" {
		t.Fatal("expected drop, got ", cfg.UI.WindowDragMode)
	}
}

type TestListsConfig struct {
	Int []int    `ini:"integers"`
	Str []string `ini:"strings"`
}

func TestLists(t *testing.T) {
	data := `
integers=1,2,3,4
strings= foo,bar,gofmt
`

	var cfg TestListsConfig
	if err := Unmarshal([]byte(data), &cfg); err != nil {
		t.Fatal(err)
	}

	totalExpected := 10
	total := 0
	for _, i := range cfg.Int {
		total += i
	}

	if totalExpected != total {
		t.Fatal("expected", totalExpected, "got", total)
	}

	if len(cfg.Str) != 3 && !slices.Contains(cfg.Str, "gofmt") {
		t.Fatal("expected slice len 3 and gofmt as element")
	}
}
