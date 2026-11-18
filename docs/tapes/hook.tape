# Spot reacting to tests in zsh: green, red, then the comeback.
Output docs/demo.gif
Set Shell zsh
Set FontSize 18
Set Width 1200
Set Height 720
Set Padding 24
Set Theme "Catppuccin Mocha"
Hide
Type "source docs/tapes/env.sh" Enter
Sleep 4s
Show
Sleep 500ms
Type@60ms "go test ./..." Sleep 300ms Enter
Sleep 2.5s
Screenshot docs/shot-pass.png
Type@60ms "BREAK=1 go test ./..." Sleep 300ms Enter
Sleep 3s
Screenshot docs/shot-fail.png
Type "clear" Enter
Sleep 300ms
Type@60ms "go test ./..." Sleep 300ms Enter
Sleep 4.5s
Screenshot docs/shot-comeback.png
Sleep 1s
