package game

import rl "github.com/gen2brain/raylib-go/raylib"

type Button struct {
	Text string
	Rect rl.Rectangle
	RectColor rl.Color
	TextColor rl.Color
	TextColorHovered rl.Color
	TextFactor float32
	Callback func()
}

var uiFont *rl.Font

func SetUIFont(font *rl.Font) {
	uiFont = font
}

func CreateButton(text string, x float32, y float32, w float32, h float32) Button {
	return Button{
		Text: text,
		Rect: rl.Rectangle{x, y, w, h},
		RectColor: rl.Red,
		TextColor: rl.White,
		TextColorHovered: rl.LightGray,
		TextFactor: 0.75,
	}
}

func (b Button) Render() {
	color := b.TextColor
	textSize := int32(b.Rect.Height*b.TextFactor)
	textX := b.Rect.X + b.Rect.Width/2 - float32(rl.MeasureText(b.Text, textSize))/2
	textY := b.Rect.Y + b.Rect.Height/2 - float32(textSize)/2

	mousePos := rl.GetMousePosition()
	mouseOverButton := (
		mousePos.X > b.Rect.X && mousePos.X < b.Rect.X + b.Rect.Width &&
		mousePos.Y > b.Rect.Y && mousePos.Y < b.Rect.Y + b.Rect.Height)

	if mouseOverButton {
		color = b.TextColorHovered
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			if b.Callback != nil {
				b.Callback()
			}
		}
		if cursor == rl.MouseCursorDefault {
			cursor = rl.MouseCursorPointingHand
		}
	}

	rl.DrawRectangleRec(b.Rect, b.RectColor)
	rl.DrawTextEx(
		*uiFont, b.Text,
		rl.Vector2{float32(textX), float32(textY)},
		float32(textSize), 4.0, color)
}

var liveInfoY float32 = 0

func LiveInfoFrameReset() {
	liveInfoY = 0
}

func LiveInfo(text string) {
	DrawText(text, 0, liveInfoY, 24, rl.White)
	liveInfoY += 24
}

func DrawText(text string, x float32, y float32, size int, color rl.Color) {
	rl.DrawTextEx(
		*uiFont, text,
		rl.Vector2{x, y},
		float32(size), 4.0, color)
}
