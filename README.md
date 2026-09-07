# encoding/ini, Golang INI Parser
This is a simple Golang package implementation inspired by encoding/json/v2.

It however pushes simplicity to the limit, relying on regex and scanning lines instead of parsing each token separatedly.

Other packages for INI files unmarshaling:
- [wlevene/ini](github.com/wlevene/ini) also has a bug when unmarshaling complex structures, such as the Config used by [kivutar/goro](https://github.com/kivutar/goro/blob/main/config/config.go#L15).
- [ncpa0cpl/ini](https://github.com/ncpa0cpl/ini) manages to unmarshal the above complex structure but would fail to preserve default values.

Unfortunately maintenance in those packages became difficult since they've been made by AI, and AI design decisions directly impacts the usability and maintainability of the package.

If needed, INI dialects could be easily implemented.

# Features

- Read by []byte
- Read from files
- Unmarshal to Struct

# Installation

```shell
go get github.com/jacksonbenete/encoding_ini
```

# Example

Given a complex ini file:
```
[login]
username=a
password=
auto_login=true

[window]
fullscreen = false

[render]
vsync = true
fps = false

[audio]
bgm_volume = 0.55
sfx_volume = 0.55

[gameplay]
no_shift = false
no_ctrl = true
less_effects = false
snap = false
itemsnap = false
```

Unmarshal to struct:
```go
func TestIniParser(t *testing.T) {
cfg := Config{}

err := ini.Unmarshal([]byte(goro_ini), &cfg)
if err != nil {
t.Fatalf("cannot unmarshal ini: %v", err)
}

fmt.Printf("Resulting struct: %+v", cfg)
}
```

Result:
```
Resulting struct: &{DataDir: Window:{Title: Width:0 Height:0 Fullscreen:false} Packet:{ClientDate:0 Profile:0} Login:{Username:a Password: AutoLogin:true CharSlot:0} Audio:{Disabled:false BGM:true BGMVolume:0.55 SFXVolume:0.55} Render:{GraphicsAPI: VSync:true FPS:false NoUI:false AsyncUI:true UIProfile:false BenchSeconds:0 BenchWarmupSeconds:0 CPUProfile: Stats:false WorldDebugStats:false} Network:{Trace:false} Fog:{Enabled:true} Gameplay:{NoShift:false NoCtrl:true LessEffects:false SnapTargets:false SnapItems:false ForceUserAI:false} Script:{Path:} Log:{Level: File:}}
```
