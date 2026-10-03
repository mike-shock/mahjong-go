package main

import (
 "fmt"
 "image/color"
 "time"

 "mahjong/internal/game"
 "fyne.io/fyne/v2"
 "fyne.io/fyne/v2/app"
 "fyne.io/fyne/v2/canvas"
 "fyne.io/fyne/v2/container"
 "fyne.io/fyne/v2/dialog"
 "fyne.io/fyne/v2/theme"
 "fyne.io/fyne/v2/widget"
)

type boardLayout struct{}
func (boardLayout) MinSize([]fyne.CanvasObject) fyne.Size {return fyne.NewSize(840,560)}
func (boardLayout) Layout(objects []fyne.CanvasObject,size fyne.Size) {
 sx,sy:=size.Width/840,size.Height/560
 for _,o:=range objects {
  t:=o.(*tileButton)
  t.Move(fyne.NewPos(float32(20+t.tile.X*64+t.tile.Z*10)*sx,float32(20+t.tile.Y*62-t.tile.Z*10)*sy))
  t.Resize(fyne.NewSize(60*sx,58*sy))
 }
}

type tileButton struct {
 widget.BaseWidget
 tile game.Tile
 label string
 selected,hinted,free bool
 tapped func()
}
func (t *tileButton) Tapped(*fyne.PointEvent) {t.tapped()}
func (t *tileButton) CreateRenderer() fyne.WidgetRenderer {
 bg:=canvas.NewRectangle(color.NRGBA{R:244,G:235,B:210,A:255})
 bg.CornerRadius=6
 bg.StrokeWidth=2
 bg.StrokeColor=color.NRGBA{R:170,G:151,B:112,A:255}
 if !t.free {bg.FillColor=color.NRGBA{R:169,G:166,B:151,A:255}}
 if t.selected {bg.FillColor=color.NRGBA{R:255,G:209,B:92,A:255}}
 if t.hinted {bg.StrokeColor=color.NRGBA{R:40,G:190,B:130,A:255};bg.StrokeWidth=4}
 text:=canvas.NewText(t.label,color.NRGBA{R:36,G:57,B:67,A:255})
 if t.tile.Kind<9 {text.Color=color.NRGBA{R:179,G:49,B:43,A:255}}
 if t.tile.Kind>=9 && t.tile.Kind<18 {text.Color=color.NRGBA{R:28,G:113,B:68,A:255}}
 text.Alignment=fyne.TextAlignCenter
 text.TextSize=17
 text.TextStyle.Bold=true
 return widget.NewSimpleRenderer(container.NewStack(bg,container.NewCenter(text)))
}
func tileName(k int) string {
 if k<9 {return fmt.Sprintf("%d 万",k+1)}
 if k<18 {return fmt.Sprintf("%d ║",k-8)}
 if k<27 {return fmt.Sprintf("%d ●",k-17)}
 return []string{"Восток","Юг","Запад","Север","Красн.","Зелён.","Белый","Цветок","Сезон"}[k-27]
}

func main() {
 a:=app.NewWithID("ru.mahjong.solitaire")
 a.Settings().SetTheme(theme.LightTheme())
 w:=a.NewWindow("Маджонг — пасьянс")
 w.Resize(fyne.NewSize(1000,720))
 g:=game.New(time.Now().UnixNano())
 selected:=-1
 hintA,hintB:=-1,-1
 status:=widget.NewLabel("")
 message:=widget.NewLabel("Выберите две одинаковые свободные плитки.")
 board:=container.New(boardLayout{})
 var refresh func()
 refresh=func() {
  board.Objects=nil
  for i,t:=range g.Tiles {
   if t.Removed {continue}
   i:=i
   button:=&tileButton{tile:t,label:tileName(t.Kind),free:g.Free(i),selected:selected==i,hinted:i==hintA||i==hintB}
   button.ExtendBaseWidget(button)
   button.tapped=func() {
    if !g.Free(i) {message.SetText("Плитка закрыта сверху или с обеих сторон.");return}
    hintA,hintB=-1,-1
    if selected==i {selected=-1} else if selected<0 {selected=i} else if g.Match(selected,i) {
     selected=-1
     message.SetText("Пара снята.")
    } else {selected=i;message.SetText("Нужны две одинаковые плитки.")}
    refresh()
   }
   board.Add(button)
  }
  status.SetText(fmt.Sprintf("Плиток: %d / 144",g.Remaining()))
  if g.Remaining()==0 {message.SetText("Поздравляем! Все плитки сняты.")} else if _,_,ok:=g.Hint();!ok {message.SetText("Доступных пар нет. Отмените ход или перемешайте плитки.")}
  board.Refresh()
 }
 newGame:=widget.NewButton("Новая игра",func(){
  dialog.ShowConfirm("Новая игра","Начать заново?",func(ok bool){if ok {g=game.New(time.Now().UnixNano());selected=-1;hintA,hintB=-1,-1;message.SetText("Выберите пару.");refresh()}},w)
 })
 hint:=widget.NewButton("Подсказка",func(){
  x,y,ok:=g.Hint()
  if ok {selected=-1;hintA,hintB=x,y;message.SetText("Пара выделена зелёной рамкой.");refresh()} else {message.SetText("Доступных пар нет.")}
 })
 undo:=widget.NewButton("Отменить ход",func(){if g.Undo(){selected=-1;hintA,hintB=-1,-1;message.SetText("Ход отменён.");refresh()}})
 shuffle:=widget.NewButton("Перемешать",func(){g.Shuffle();selected=-1;hintA,hintB=-1,-1;message.SetText("Плитки перемешаны. История ходов очищена.");refresh()})
 rules:=widget.NewButton("Правила",func(){dialog.ShowInformation("Как играть","Снимайте пары одинаковых плиток. Плитка свободна, если над ней ничего нет и хотя бы одна боковая сторона открыта.\n\nЦель — убрать все 144 плитки. Цветок и сезон сопоставляются только с такими же плитками. Новая раскладка и перемешивание имеют решение, но выбранные ходы могут привести в тупик.",w)})
 background:=canvas.NewRectangle(color.NRGBA{R:24,G:77,B:66,A:255})
 header:=container.NewVBox(widget.NewLabelWithStyle("МАДЖОНГ",fyne.TextAlignCenter,fyne.TextStyle{Bold:true}),container.NewHBox(newGame,hint,undo,shuffle,rules,status))
 w.SetContent(container.NewBorder(header,message,nil,nil,container.NewStack(background,board)))
 refresh()
 w.ShowAndRun()
}
