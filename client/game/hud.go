package game

import (
	"fmt"
	stdimage "image"
	"image/color"
	"sort"
	"strconv"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/unit/ability"
	"github.com/xescugc/maze-wars/utils"
)

const (
	unitToolTipTmpl       = "Lvl: %d\nGold: %d\nHP: %.0f\nSpeed: %.f\nIncome: %d\nEnv: %s\nKeybind: %s"
	unitUpdateToolTipTmpl = "Lvl: %d\nCost: %d\nGold: %d\nHP: %.0f\nIncome: %d"

	unitAttackToolTipTmpl       = "Lvl: %d\nGold: %d\nHP: %.0f\nDamage: %.0f\nAttack Speed: %.0f\nSpeed: %.f\nIncome: %d\nEnv: %s\nKeybind: %s"
	unitAttackUpdateToolTipTmpl = "Lvl: %d\nCost: %d\nGold: %d\nHP: %.0f\nDamage: %.0f\nIncome: %d"

	towerRemoveToolTipTmpl = "Selling the tower will give back %d gold"
	towerUpdateToolTipTmpl = "Cost: %d\nDamage: %.0f\nAttack Speed: %.0f\nHealth: %.0f\nKeybind: %s"
	towerUpdateLimit       = "Tower is at it's max level"

	unitToolTipTitleTmpl = "%s (%s)"

	isPressed = true
)

type btnWithGraphic struct {
	btn      *widget.Button
	enabled  *widget.Graphic
	disabled *widget.Graphic
}

type unitTooltips struct {
	title              *widget.Text
	currentLvl         *widget.Text
	nextLvl            *widget.Text
	gold               *widget.Text
	ngold              *widget.Text
	ngoldDiff          *widget.Text
	income             *widget.Text
	nincome            *widget.Text
	nincomeDiff        *widget.Text
	health             *widget.Text
	nhealth            *widget.Text
	nhealthDiff        *widget.Text
	movementSpeed      *widget.Text
	nmovementSpeed     *widget.Text
	nmovementSpeedDiff *widget.Text
}

type towerTooltips struct {
}

// HUDStore is in charge of keeping track of all the elements
// on the player HUD that are static and always seen
type HUDStore struct {
	*flux.ReduceStore

	game *Game

	ui *ebitenui.UI

	//statsListW   *widget.List
	//incomeTextW  *widget.Text
	//winLoseTextW *widget.Text

	//unitsC       *widget.Container
	unitsTooltip  map[string]*unitTooltips
	towerTooltips map[string]*towerTooltips

	//unitUpdatesC       *widget.Container
	//unitUpdatesTooltip map[string]*widget.Text

	//towersC *widget.Container

	//bottomLeftContainer *widget.Container
	//towerMenuContainer  *widget.Container
	//towerRemoveToolTip  *widget.Text
	//towerUpdateToolTip1 *widget.Text
	//towerUpdateButton1  *widget.Button
	//towerUpdateToolTip2 *widget.Text
	//towerUpdateButton2  *widget.Button

	// New

	displayDefaultC *widget.Container

	displayTargetC         *widget.Container
	displayTargetProfile   *widget.Graphic
	displayTargetHealth    *widget.Text
	displayTargetHealthBar *widget.ProgressBar

	// Units Target
	displayTargetUnitC                 *widget.Container
	displayTargetUnitNameTxtW          *widget.Text
	displayTargetUnitMovementSpeedTxtW *widget.Text
	displayTargetUnitBountyTxtW        *widget.Text
	displayTargetUnitPlayerTxtW        *widget.Text
	displayTargetUnitAbilityImage1     *widget.Graphic

	// Towers Target
	displayTargetTowerC                            *widget.Container
	displayTargetTowerUpdateC1                     *widget.Container
	displayTargetTowerUpdateButton1                *widget.Button
	displayTargetTowerUpdateImage1                 *widget.Graphic
	displayTargetTowerUpdateToolTip1TitleTxt       *widget.Text
	displayTargetTowerUpdateToolTip1GoldTxt        *widget.Text
	displayTargetTowerUpdateToolTip1DamageTxt      *widget.Text
	displayTargetTowerUpdateToolTip1RangeTxt       *widget.Text
	displayTargetTowerUpdateToolTip1HealthTxt      *widget.Text
	displayTargetTowerUpdateToolTip1DescriptionTxt *widget.Text
	displayTargetTowerUpdateC2                     *widget.Container
	displayTargetTowerUpdateButton2                *widget.Button
	displayTargetTowerUpdateImage2                 *widget.Graphic
	displayTargetTowerUpdateToolTip2TitleTxt       *widget.Text
	displayTargetTowerUpdateToolTip2GoldTxt        *widget.Text
	displayTargetTowerUpdateToolTip2DamageTxt      *widget.Text
	displayTargetTowerUpdateToolTip2RangeTxt       *widget.Text
	displayTargetTowerUpdateToolTip2HealthTxt      *widget.Text
	displayTargetTowerUpdateToolTip2DescriptionTxt *widget.Text
	displayTargetTowerSellToolTip                  *widget.Text
	displayTargetTowerRangeTxtW                    *widget.Text
	displayTargetTowerDamageTxtW                   *widget.Text
	displayTargetTowerAttackSpeedTxtW              *widget.Text
	displayTargetTowerNameTxtW                     *widget.Text

	towersGC *widget.Container

	unitsGC              *widget.Container
	unitsBtns            map[string]*btnWithGraphic
	unitsUpdateAnimation map[string][]*widget.Graphic
	unitAnimationCount   int

	scoreboardW     *widget.Window
	scoreboardC     *widget.Container
	scoreboardBodyC *widget.Container

	infoTimerTxt       *widget.Text
	infoGoldTxt        *widget.Text
	infoCapTxt         *widget.Text
	infoLivesTxt       *widget.Text
	infoIncomeTxt      *widget.Text
	infoIncomeTimerTxt *widget.Text

	menuBtnW  *widget.Button
	menuW     *widget.Window
	keybindsW *widget.Window
}

// HUDState stores the HUD state
type HUDState struct {
	SelectedTower *SelectedTower
	OpenTowerMenu *store.Tower
	OpenUnitMenu  *store.Unit

	LastCursorPosition utils.Object

	ShowStats      bool
	ShowScoreboard bool
}

type SelectedTower struct {
	store.Tower

	Invalid bool
}

var (
	// The key value of this maps is the TYPE of the Unit|Tower
	unitKeybinds  = make(map[string]ebiten.Key)
	towerKeybinds = make(map[string]ebiten.Key)

	sellTowerKeybind = ebiten.KeyD

	updateTowerKeybind1 = ebiten.KeyZ
	updateTowerKeybind2 = ebiten.KeyX

	rangeTowerKeybind = ebiten.KeyQ
	meleeTowerKeybind = ebiten.KeyW
)

func init() {
	for _, u := range unit.Units {
		var k ebiten.Key
		err := k.UnmarshalText([]byte(u.Keybind))
		if err != nil {
			panic(err)
		}
		unitKeybinds[u.Type.String()] = k
	}

	towerKeybinds[tower.Range1.String()] = rangeTowerKeybind
	towerKeybinds[tower.Melee1.String()] = meleeTowerKeybind
}

// NewHUDStore creates a new HUDStore with the Dispatcher d and the Game g
func NewHUDStore(d *flux.Dispatcher, g *Game) (*HUDStore, error) {
	hs := &HUDStore{
		game:          g,
		unitsTooltip:  make(map[string]*unitTooltips),
		towerTooltips: make(map[string]*towerTooltips),
		//unitUpdatesTooltip:   make(map[string]*widget.Text),
		unitsBtns:            make(map[string]*btnWithGraphic),
		unitsUpdateAnimation: make(map[string][]*widget.Graphic),
	}
	hs.ReduceStore = flux.NewReduceStore(d, hs.Reduce, HUDState{
		ShowStats:      true,
		ShowScoreboard: false,
	})

	hs.buildUI()

	return hs, nil
}

func (hs *HUDStore) validateOpenTower(tw *store.Tower, tws map[string]*store.Tower) {
	if tw != nil {
		if tw != nil {
			if _, ok := tws[tw.ID]; !ok {
				actionDispatcher.CloseTowerMenu()
			}
		}
	}
}

func (hs *HUDStore) validateOpenUnit(ut *store.Unit, uts map[string]*store.Unit) {
	if ut != nil {
		if ut != nil {
			if _, ok := uts[ut.ID]; !ok {
				actionDispatcher.CloseUnitMenu()
			}
		}
	}
}

func (hs *HUDStore) Update() error {
	b := time.Now()
	defer utils.LogTime(hs.game.Logger, b, "hud update")

	hs.ui.Update()

	cs := hs.game.Camera.GetState().(CameraState)
	hst := hs.GetState().(HUDState)
	x, y := ebiten.CursorPosition()
	cp := hs.game.Store.Lines.FindCurrentPlayer()
	cl := hs.game.Store.Lines.FindLineByID(cp.LineID)
	tws := cl.Towers
	uts := cl.Units
	// Only send a CursorMove when the curso has actually moved
	if int(hst.LastCursorPosition.X) != x || int(hst.LastCursorPosition.Y) != y {
		actionDispatcher.CursorMove(x, y)
	}
	// If the Current player is dead or has no more lives there are no
	// mo actions that can be done
	if cp.Lives == 0 || cp.Winner {
		return nil
	}

	hs.validateOpenTower(hst.OpenTowerMenu, tws)
	hs.validateOpenUnit(hst.OpenUnitMenu, uts)
	// As the current opened tower
	if ebiten.IsKeyPressed(ebiten.KeyTab) && !hst.ShowScoreboard {
		actionDispatcher.ShowScoreboard(true)
	} else if !ebiten.IsKeyPressed(ebiten.KeyTab) && hst.ShowScoreboard {
		actionDispatcher.ShowScoreboard(false)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		clickAbsolute := utils.Object{
			X: float64(x) + cs.X,
			Y: float64(y) + cs.Y,
			W: 1, H: 1,
		}

		if hst.SelectedTower != nil && !hst.SelectedTower.Invalid {
			actionDispatcher.PlaceTower(hst.SelectedTower.Type, cp.ID, int(hst.SelectedTower.X+cs.X), int(hst.SelectedTower.Y+cs.Y))
			return nil
		}
		for _, t := range tws {
			if clickAbsolute.IsColliding(t.Object) && cp.ID == t.PlayerID {
				actionDispatcher.OpenTowerMenu(t.ID)
				return nil
			}
		}
		for _, u := range uts {
			if clickAbsolute.IsColliding(u.Object) {
				actionDispatcher.OpenUnitMenu(u.ID)
				return nil
			}
		}
		// If we are here no Tower was clicked but a click action was done,
		// so we check if the OpenTowerMenu is set to unset it as this was
		// a click-off
		if hst.OpenTowerMenu != nil || hst.OpenUnitMenu != nil {
			p := stdimage.Point{x, y}
			w := hs.displayTargetC.GetWidget()
			inside := p.In(w.Rect)
			layer := w.EffectiveInputLayer()

			clickedInside := inside && input.MouseButtonPressedLayer(ebiten.MouseButtonLeft, layer)

			if !clickedInside {
				actionDispatcher.CloseTowerMenu()
				actionDispatcher.CloseUnitMenu()
			}
		}
	}

	for ut, kb := range unitKeybinds {
		if cp.CanSummonUnit(ut) && inpututil.IsKeyJustPressed(kb) {
			hs.unitsBtns[ut].btn.Click()
			return nil
		}
	}
	for tt, kb := range towerKeybinds {
		if cp.CanPlaceTower(tt) && inpututil.IsKeyJustPressed(kb) {
			actionDispatcher.SelectTower(tt, x, y)
			return nil
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		actionDispatcher.GoHome()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if hst.OpenTowerMenu != nil || hst.OpenUnitMenu != nil {
			if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				actionDispatcher.CloseTowerMenu()
				actionDispatcher.CloseUnitMenu()
			}
		} else {
			if hs.ui.IsWindowOpen(hs.menuW) {
				hs.closeModal(hs.menuW)
			} else if hs.ui.IsWindowOpen(hs.keybindsW) {
				hs.closeModal(hs.keybindsW)
			} else {
				hs.menuBtnW.Click()
			}
		}
	}
	if hst.OpenTowerMenu != nil {
		if inpututil.IsKeyJustPressed(sellTowerKeybind) {
			actionDispatcher.RemoveTower(cp.ID, hst.OpenTowerMenu.ID)
			actionDispatcher.CloseTowerMenu()
		}
		tw := tower.Towers[hst.OpenTowerMenu.Type]
		if len(tw.Updates) >= 1 && inpututil.IsKeyJustPressed(updateTowerKeybind1) {
			actionDispatcher.UpdateTower(cp.ID, hst.OpenTowerMenu.ID, tw.Updates[0].String())
		}
		if len(tw.Updates) >= 2 && inpututil.IsKeyJustPressed(updateTowerKeybind2) {
			actionDispatcher.UpdateTower(cp.ID, hst.OpenTowerMenu.ID, tw.Updates[1].String())
		}
	}
	if hst.SelectedTower != nil {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			actionDispatcher.DeselectTower(hst.SelectedTower.Type)
		} else {
			invalid := !cp.CanPlaceTower(hst.SelectedTower.Type)

			neo := hst.SelectedTower.Object
			neo.X += cs.X
			neo.Y += cs.Y

			if !invalid {
				invalid = !cl.Graph.CanAddTower(int(neo.X), int(neo.Y), neo.W, neo.H)
			}

			if !invalid {
				for _, u := range cl.Units {
					if u.IsColliding(neo) {
						invalid = true
						break
					}
				}
			}

			if invalid != hst.SelectedTower.Invalid {
				actionDispatcher.SelectedTowerInvalid(invalid)
			}
		}
	}

	return nil
}

func (hs *HUDStore) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(hs.game.Logger, b, "hud draw")

	hst := hs.GetState().(HUDState)
	cs := hs.game.Camera.GetState().(CameraState)
	cp := hs.game.Store.Lines.FindCurrentPlayer()
	cl := hs.game.Store.Lines.FindLineByID(cp.LineID)

	hs.validateOpenTower(hst.OpenTowerMenu, cl.Towers)
	hs.validateOpenUnit(hst.OpenUnitMenu, cl.Units)

	if hst.ShowScoreboard {
		hs.buildScoreboard()
		hs.loadModal(hs.scoreboardW)
	} else {
		hs.closeModal(hs.scoreboardW)
	}

	psit := hs.game.Store.Lines.GetIncomeTimer()
	//entries := make([]any, 0, len(hs.game.Store.Lines.ListPlayers())+1)
	//entries = append(entries,
	//cutils.ListEntry{
	//Text: fmt.Sprintf("%s %s %s",
	//cutils.FillIn("Name", 10),
	//cutils.FillIn("Lives", 8),
	//cutils.FillIn("Income", 8)),
	//},
	//)

	//var sortedPlayers = make([]*store.Player, 0, 0)
	//for _, p := range hs.game.Store.Lines.ListPlayers() {
	//sortedPlayers = append(sortedPlayers, p)
	//}
	//sort.Slice(sortedPlayers, func(i, j int) bool {
	//ii := sortedPlayers[i]
	//jj := sortedPlayers[j]
	//if ii.Income != jj.Income {
	//return ii.Income > jj.Income
	//}
	//return ii.LineID < jj.LineID
	//})
	//for _, p := range sortedPlayers {
	//entries = append(entries,
	//cutils.ListEntry{
	//ID: p.ID,
	//Text: fmt.Sprintf("%s %s %s",
	//cutils.FillIn(p.Name, 10),
	//cutils.FillIn(strconv.Itoa(p.Lives), 8),
	//cutils.FillIn(strconv.Itoa(p.Income), 8)),
	//},
	//)
	//}
	//if !cutils.EqualListEntries(entries, hs.statsListW.Entries().([]any)) {
	//hs.statsListW.SetEntries(entries)
	//}

	//visibility := widget.Visibility_Show
	//if !hst.ShowStats {
	//visibility = widget.Visibility_Hide_Blocking
	//}
	//hs.statsListW.GetWidget().Visibility = visibility
	//hs.incomeTextW.Label = fmt.Sprintf("Gold: %s Income Timer: %ds", cutils.FillIn(strconv.Itoa(cp.Gold), 5), psit)

	hs.infoTimerTxt.Label = cutils.FmtDuration(time.Now().Sub(hs.game.Store.Lines.GetStartedAt()))
	hs.infoGoldTxt.Label = strconv.Itoa(cp.Gold)
	hs.infoCapTxt.Label = fmt.Sprintf("%d/%d", cp.Capacity, utils.MaxCapacity)
	hs.infoLivesTxt.Label = strconv.Itoa(cp.Lives)
	hs.infoIncomeTxt.Label = strconv.Itoa(cp.Income)
	hs.infoIncomeTimerTxt.Label = fmt.Sprintf("%ds", psit)

	//wuts := hs.unitsC.Children()
	for _, u := range sortedUnits() {
		uu := cp.UnitUpdates[u.Type.String()]
		//wuts[i].GetWidget().Disabled = !cp.CanSummonUnit(u.Type.String())
		// NEW
		cpcs := cp.CanSummonUnit(u.Type.String())
		hs.unitsBtns[u.Type.String()].btn.GetWidget().Disabled = !cpcs
		hs.unitsBtns[u.Type.String()].enabled.GetWidget().Visibility = widget.Visibility_Show
		hs.unitsBtns[u.Type.String()].disabled.GetWidget().Visibility = widget.Visibility_Hide
		if !cpcs {
			hs.unitsBtns[u.Type.String()].enabled.GetWidget().Visibility = widget.Visibility_Hide
			hs.unitsBtns[u.Type.String()].disabled.GetWidget().Visibility = widget.Visibility_Show
		}
		hs.unitsTooltip[u.Type.String()].title.Label = fmt.Sprintf(unitToolTipTitleTmpl, u.Name(), u.Keybind)
		hs.unitsTooltip[u.Type.String()].currentLvl.Label = fmt.Sprint(uu.Level)
		hs.unitsTooltip[u.Type.String()].nextLvl.Label = fmt.Sprintf("%d (cost %d)", uu.Level+1, uu.UpdateCost)
		//hs.unitsTooltip[u.Type.String()].ucost.Label = fmt.Sprintf("lvl %d (cost %d)", uu.Level+1, uu.UpdateCost)
		//hs.unitsTooltip[u.Type.String()].ucost.Label = fmt.Sprint(uu.UpdateCost)
		hs.unitsTooltip[u.Type.String()].gold.Label = fmt.Sprint(uu.Current.Gold)
		hs.unitsTooltip[u.Type.String()].ngold.Label = fmt.Sprintf("%d", uu.Next.Gold)
		hs.unitsTooltip[u.Type.String()].ngoldDiff.Label = fmt.Sprintf("(+%d)", uu.Next.Gold-uu.Current.Gold)
		hs.unitsTooltip[u.Type.String()].income.Label = fmt.Sprint(uu.Current.Income)
		hs.unitsTooltip[u.Type.String()].nincome.Label = fmt.Sprintf("%d", uu.Next.Income)
		hs.unitsTooltip[u.Type.String()].nincomeDiff.Label = fmt.Sprintf(" (+%d)", uu.Next.Income-uu.Current.Income)
		hs.unitsTooltip[u.Type.String()].health.Label = fmt.Sprint(uu.Current.Health)
		hs.unitsTooltip[u.Type.String()].nhealth.Label = fmt.Sprintf("%.0f", uu.Next.Health)
		hs.unitsTooltip[u.Type.String()].nhealthDiff.Label = fmt.Sprintf("(+%.0f)", uu.Next.Health-uu.Current.Health)
		hs.unitsTooltip[u.Type.String()].movementSpeed.Label = fmt.Sprint(uu.Current.MovementSpeed)
		hs.unitsTooltip[u.Type.String()].nmovementSpeed.Label = fmt.Sprintf("%.0f", uu.Next.MovementSpeed)
		hs.unitsTooltip[u.Type.String()].nmovementSpeedDiff.Label = fmt.Sprintf("(+%.0f)", uu.Next.MovementSpeed-uu.Current.MovementSpeed)
		// END NEW
		//if u.HasAbility(ability.Attack) {
		//hs.unitsTooltip[u.Type.String()].Label = fmt.Sprintf(unitAttackToolTipTmpl, uu.Level, uu.Current.Gold, uu.Current.Health, uu.Current.Damage, uu.Current.AttackSpeed, uu.Current.MovementSpeed, uu.Current.Income, u.Environment, u.Keybind)
		//} else {
		//hs.unitsTooltip[u.Type.String()].Label = fmt.Sprintf(unitToolTipTmpl, uu.Level, uu.Current.Gold, uu.Current.Health, uu.Current.MovementSpeed, uu.Current.Income, u.Environment, u.Keybind)
		//}
	}

	// TODO: Add the Upgrade display
	//wuuts := hs.unitUpdatesC.Children()
	for _, u := range sortedUnits() {
		//uu := cp.UnitUpdates[u.Type.String()]
		//wuuts[i].GetWidget().Disabled = !cp.CanUpdateUnit(u.Type.String())
		//if u.HasAbility(ability.Attack) {
		//hs.unitUpdatesTooltip[u.Type.String()].Label = fmt.Sprintf(unitAttackUpdateToolTipTmpl, uu.Level+1, uu.UpdateCost, uu.Next.Gold, uu.Next.Health, uu.Next.Damage, uu.Next.Income)
		//} else {
		//hs.unitUpdatesTooltip[u.Type.String()].Label = fmt.Sprintf(unitUpdateToolTipTmpl, uu.Level+1, uu.UpdateCost, uu.Next.Gold, uu.Next.Health, uu.Next.Income)
		//}
		if cp.CanUpdateUnit(u.Type.String()) {
			ic := (hs.unitAnimationCount / 15) % 4
			for i, a := range hs.unitsUpdateAnimation[u.Type.String()] {
				a.GetWidget().Visibility = widget.Visibility_Hide
				if i == ic {
					a.GetWidget().Visibility = widget.Visibility_Show
				}
			}
			// TODO: Maybe set it to 0 and do not do this operation here?
			// potentially have a global counter for animations
			hs.unitAnimationCount += 1
		} else {
			for _, a := range hs.unitsUpdateAnimation[u.Type.String()] {
				a.GetWidget().Visibility = widget.Visibility_Hide
			}
		}
	}

	//wtws := hs.towersC.Children()
	//for i, t := range sortedTowers() {
	//wtws[i].GetWidget().Disabled = !cp.CanPlaceTower(t.Type.String())
	//}

	// TODO: Fix this
	//if cp.Lives == 0 {
	//hs.winLoseTextW.Label = "YOU LOST"
	//hs.winLoseTextW.GetWidget().Visibility = widget.Visibility_Show
	//} else if cp.Winner {
	//hs.winLoseTextW.Label = "YOU WON!"
	//hs.winLoseTextW.GetWidget().Visibility = widget.Visibility_Show
	//} else if hs.winLoseTextW.GetWidget().Visibility != widget.Visibility_Hide {
	//hs.winLoseTextW.Label = ""
	//hs.winLoseTextW.GetWidget().Visibility = widget.Visibility_Hide
	//}

	hs.displayTargetC.GetWidget().Visibility = widget.Visibility_Hide
	hs.displayDefaultC.GetWidget().Visibility = widget.Visibility_Show
	if hst.OpenTowerMenu != nil {
		hs.displayTargetC.GetWidget().Visibility = widget.Visibility_Show
		hs.displayTargetTowerC.GetWidget().Visibility = widget.Visibility_Show

		hs.displayTargetUnitC.GetWidget().Visibility = widget.Visibility_Hide
		hs.displayDefaultC.GetWidget().Visibility = widget.Visibility_Hide

		ot := tower.Towers[hst.OpenTowerMenu.Type]
		ct := cl.Towers[hst.OpenTowerMenu.ID]

		hs.displayTargetTowerRangeTxtW.Label = fmt.Sprint(ot.Range)
		hs.displayTargetTowerDamageTxtW.Label = fmt.Sprint(ot.Damage)
		hs.displayTargetTowerAttackSpeedTxtW.Label = fmt.Sprint(ot.AttackSpeed)
		hs.displayTargetTowerNameTxtW.Label = ot.Name()
		// TODO: Add the keybind to the tooltip
		//hs.bottomLeftContainer.GetWidget().Visibility = widget.Visibility_Show
		// TODO: Fix this by being the total amount to reach here and let it be 75%
		sellTowerGoldReturn := ot.Gold / 2

		hs.displayTargetProfile.Image = cutils.Images.Get(ct.ProfileKey())
		hs.displayTargetHealth.Label = fmt.Sprintf("%0.f/%0.f", ct.Health, ot.Health)
		hs.displayTargetHealthBar.Min = 0
		hs.displayTargetHealthBar.Max = int(ot.Health)
		hs.displayTargetHealthBar.SetCurrent(int(ct.Health))

		// Where there is no more updates we disable the button
		tu := tower.Towers[hst.OpenTowerMenu.Type].Updates
		if len(tu) == 0 {
			//hs.towerUpdateButton1.GetWidget().Visibility = widget.Visibility_Hide
			//hs.towerUpdateButton2.GetWidget().Visibility = widget.Visibility_Hide

			// New
			hs.displayTargetTowerUpdateC1.GetWidget().Visibility = widget.Visibility_Hide
			hs.displayTargetTowerUpdateC2.GetWidget().Visibility = widget.Visibility_Hide
			// END New
		} else {
			if len(tu) >= 1 {
				tw := tower.Towers[tu[0].String()]
				//hs.towerUpdateButton1.Image = cutils.ButtonImageFromImage(cutils.Images.Get(tw.FacesetKey()))
				//hs.towerUpdateButton1.GetWidget().Visibility = widget.Visibility_Show
				//hs.towerUpdateButton1.GetWidget().Disabled = cp.Gold < tw.Gold
				//hs.towerUpdateToolTip1.Label = fmt.Sprintf(towerUpdateToolTipTmpl, tw.Gold, tw.Damage, tw.AttackSpeed, tw.Health, updateTowerKeybind1)
				//hs.towerUpdateButton2.GetWidget().Visibility = widget.Visibility_Hide

				// New
				hs.displayTargetTowerUpdateImage1.Image = cutils.Images.Get(tw.FacesetKey())
				hs.displayTargetTowerUpdateC1.GetWidget().Visibility = widget.Visibility_Show
				hs.displayTargetTowerUpdateButton1.GetWidget().Disabled = cp.Gold < tw.Gold
				hs.displayTargetTowerUpdateToolTip1TitleTxt.Label = fmt.Sprintf(unitToolTipTitleTmpl, tw.Name(), updateTowerKeybind1)
				hs.displayTargetTowerUpdateToolTip1GoldTxt.Label = fmt.Sprint(tw.Gold)
				hs.displayTargetTowerUpdateToolTip1DamageTxt.Label = fmt.Sprint(tw.Damage)
				hs.displayTargetTowerUpdateToolTip1RangeTxt.Label = fmt.Sprint(tw.Range)
				hs.displayTargetTowerUpdateToolTip1HealthTxt.Label = fmt.Sprint(tw.Health)
				hs.displayTargetTowerUpdateToolTip1DescriptionTxt.Label = tw.Description()
				hs.displayTargetTowerUpdateC2.GetWidget().Visibility = widget.Visibility_Hide
				// END New
			}
			if len(tu) >= 2 {
				tw := tower.Towers[tu[1].String()]
				//hs.towerUpdateButton2.Image = cutils.ButtonImageFromImage(cutils.Images.Get(tw.FacesetKey()))
				//hs.towerUpdateButton2.GetWidget().Visibility = widget.Visibility_Show
				//hs.towerUpdateButton2.GetWidget().Disabled = cp.Gold < tw.Gold
				//hs.towerUpdateToolTip2.Label = fmt.Sprintf(towerUpdateToolTipTmpl, tw.Gold, tw.Damage, tw.AttackSpeed, tw.Health, updateTowerKeybind2)
				// New
				hs.displayTargetTowerUpdateImage2.Image = cutils.Images.Get(tw.FacesetKey())
				hs.displayTargetTowerUpdateC2.GetWidget().Visibility = widget.Visibility_Show
				hs.displayTargetTowerUpdateButton2.GetWidget().Disabled = cp.Gold < tw.Gold
				hs.displayTargetTowerUpdateToolTip2TitleTxt.Label = fmt.Sprintf(unitToolTipTitleTmpl, tw.Name(), updateTowerKeybind2)
				hs.displayTargetTowerUpdateToolTip2GoldTxt.Label = fmt.Sprint(tw.Gold)
				hs.displayTargetTowerUpdateToolTip2DamageTxt.Label = fmt.Sprint(tw.Damage)
				hs.displayTargetTowerUpdateToolTip2RangeTxt.Label = fmt.Sprint(tw.Range)
				hs.displayTargetTowerUpdateToolTip2HealthTxt.Label = fmt.Sprint(tw.Health)
				hs.displayTargetTowerUpdateToolTip2DescriptionTxt.Label = tw.Description()
				// END New
			}
		}
		//hs.towerRemoveToolTip.Label = fmt.Sprintf(towerRemoveToolTipTmpl, sellTowerGoldReturn, sellTowerKeybind)
		// NOTE: New
		hs.displayTargetTowerSellToolTip.Label = fmt.Sprintf(towerRemoveToolTipTmpl, sellTowerGoldReturn)

	} else if hst.OpenUnitMenu != nil {
		ou := unit.Units[hst.OpenUnitMenu.Type]
		cou := cl.Units[hst.OpenUnitMenu.ID]

		hs.displayTargetUnitC.GetWidget().Visibility = widget.Visibility_Show
		hs.displayTargetTowerC.GetWidget().Visibility = widget.Visibility_Hide

		hs.displayTargetProfile.Image = cutils.Images.Get(ou.ProfileKey())
		hs.displayTargetUnitNameTxtW.Label = fmt.Sprintf("%s (lvl %d)", ou.Name(), cou.Level)
		hs.displayTargetUnitMovementSpeedTxtW.Label = fmt.Sprint(cou.MovementSpeed)
		hs.displayTargetUnitBountyTxtW.Label = fmt.Sprint(cou.Bounty)
		ucp := hs.game.Store.Lines.FindPlayerByID(cou.PlayerID)
		hs.displayTargetUnitPlayerTxtW.Label = ucp.Name
		hs.displayTargetUnitAbilityImage1.Image = cutils.Images.Get(ability.Key(ou.Abilities[0].String()))

		hs.displayTargetHealth.Label = fmt.Sprintf("%0.f/%0.f", cou.Health, cou.MaxHealth)
		hs.displayTargetHealthBar.Min = 0
		hs.displayTargetHealthBar.Max = int(cou.MaxHealth)
		hs.displayTargetHealthBar.SetCurrent(int(cou.Health))

		hs.displayTargetC.GetWidget().Visibility = widget.Visibility_Show
		hs.displayDefaultC.GetWidget().Visibility = widget.Visibility_Hide
	} else {
		//hs.bottomLeftContainer.GetWidget().Visibility = widget.Visibility_Hide
	}

	hs.ui.Draw(screen)

	if hst.SelectedTower != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(hst.SelectedTower.X)/cs.Zoom, float64(hst.SelectedTower.Y)/cs.Zoom)
		op.GeoM.Scale(cs.Zoom, cs.Zoom)

		if hst.SelectedTower != nil && hst.SelectedTower.Invalid {
			op.ColorM.Scale(2, 0.5, 0.5, 0.9)
		}

		screen.DrawImage(cutils.Images.Get(hst.SelectedTower.IdleKey()), op)
	}
}

func (hs *HUDStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	hstate, ok := state.(HUDState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.SelectTower:
		cp := hs.game.Store.Lines.FindCurrentPlayer()
		cs := hs.game.Camera.GetState().(CameraState)
		x, y := fixPosition(cs, act.SelectTower.X, act.SelectTower.Y)
		hstate.SelectedTower = &SelectedTower{
			Tower: store.Tower{
				Object: utils.Object{
					// The Buttons have 16*2 so we want to place it on the middle so just 16
					X: x,
					Y: y,
					W: 32,
					H: 32,
				},
				Type:   act.SelectTower.Type,
				LineID: cp.LineID,
			},
		}
	case action.CursorMove:
		// We update the last seen cursor position to not resend unnecessary events
		nx := act.CursorMove.X
		ny := act.CursorMove.Y

		hstate.LastCursorPosition.X = float64(nx)
		hstate.LastCursorPosition.Y = float64(ny)

		if hstate.SelectedTower != nil {
			cs := hs.game.Camera.GetState().(CameraState)

			hstate.SelectedTower.X, hstate.SelectedTower.Y = fixPosition(cs, nx, ny)
		}
	case action.PlaceTower, action.DeselectTower:
		hstate.SelectedTower = nil
	case action.SelectedTowerInvalid:
		if hstate.SelectedTower != nil {
			hstate.SelectedTower.Invalid = act.SelectedTowerInvalid.Invalid
		}
	case action.OpenTowerMenu:
		hstate.OpenTowerMenu = hs.findTowerByID(act.OpenTowerMenu.TowerID)
		hstate.OpenUnitMenu = nil
	case action.OpenUnitMenu:
		hstate.OpenUnitMenu = hs.findUnitByID(act.OpenUnitMenu.UnitID)
		hstate.OpenTowerMenu = nil
	case action.UpdateTower:
		hs.GetDispatcher().WaitFor(hs.game.Store.Lines.GetDispatcherToken())

		// As the UpdateTower is done we need to update the OpenTowerMenu
		// so we can display the new information
		hstate.OpenTowerMenu = hs.findTowerByID(act.UpdateTower.TowerID)
	case action.CloseTowerMenu:
		hstate.OpenTowerMenu = nil
	case action.CloseUnitMenu:
		hstate.OpenUnitMenu = nil
	//case action.ToggleStats:
	//hstate.ShowStats = !hstate.ShowStats
	case action.ShowScoreboard:
		hstate.ShowScoreboard = act.ShowScoreboard.Display
	default:
	}

	return hstate
}

func (hs *HUDStore) findTowerByID(tid string) *store.Tower {
	for _, l := range hs.game.Store.Lines.ListLines() {
		if t, ok := l.Towers[tid]; ok {
			return t
		}
	}
	return nil
}

func (hs *HUDStore) findUnitByID(uid string) *store.Unit {
	for _, l := range hs.game.Store.Lines.ListLines() {
		if u, ok := l.Units[uid]; ok {
			return u
		}
	}
	return nil
}

func fixPosition(cs CameraState, x, y int) (float64, float64) {
	absnx := x + int(cs.X)
	absny := y + int(cs.Y)
	// We find the closes multiple in case the cursor moves too fast, between FPS reloads,
	// and lands in a position not 'multiple' which means the position of the SelectedTower
	// is not updated and the result is the cursor far away from the Drawing of the SelectedTower
	// as it has stayed on the previous position
	var multiple int = 16
	// If it's == 0 means it's exact but as we want to center it we remove 16 (towers are 32)
	// If it's !=0 then we find what's the remaning for
	if absnx%multiple == 0 {
		x -= 16
	} else {
		x = utils.ClosestMultiple(absnx, multiple) - 16 - int(cs.X)
	}
	if absny%multiple == 0 {
		y -= 16
	} else {
		y = utils.ClosestMultiple(absny, multiple) - 16 - int(cs.Y)
	}

	return float64(x), float64(y)
}

func sortedUnits() []*unit.Unit {
	us := make([]*unit.Unit, 0, 0)
	for _, u := range unit.TypeStrings() {
		us = append(us, unit.Units[u])
	}
	return us
}

func sortedTowers() []*tower.Tower {
	return tower.FirstTowers
}

func (hs *HUDStore) buildUI() {
	//cp := hs.game.Store.Lines.FindCurrentPlayer()
	//topRightContainer := widget.NewContainer(
	//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	//)

	//topRightVerticalRowC := widget.NewContainer(
	//widget.ContainerOpts.Layout(widget.NewRowLayout(
	//widget.RowLayoutOpts.Direction(widget.DirectionVertical),
	//widget.RowLayoutOpts.Spacing(20),
	//)),
	//widget.ContainerOpts.WidgetOpts(
	//widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
	//HorizontalPosition: widget.AnchorLayoutPositionEnd,
	//VerticalPosition:   widget.AnchorLayoutPositionStart,
	//}),
	//),
	//)

	//topRightVerticalRowWraperC := widget.NewContainer(
	//widget.ContainerOpts.Layout(widget.NewRowLayout(
	//widget.RowLayoutOpts.Direction(widget.DirectionVertical),
	//)),
	//widget.ContainerOpts.WidgetOpts(
	//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
	//Stretch: true,
	//}),
	//),
	//)

	//topRightHorizontalRowC := widget.NewContainer(
	//widget.ContainerOpts.Layout(widget.NewRowLayout(
	//widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
	//widget.RowLayoutOpts.Spacing(20),
	//)),
	//widget.ContainerOpts.WidgetOpts(
	//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
	//Position: widget.RowLayoutPositionEnd,
	//}),
	//),
	//)

	//homeBtnW := widget.NewButton(
	//widget.ButtonOpts.Image(cutils.ButtonImage),

	//widget.ButtonOpts.Text("HOME(F1)", cutils.SmallFont, &widget.ButtonTextColor{
	//Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
	//}),

	//// specify that the button's text needs some padding for correct display
	//widget.ButtonOpts.TextPadding(widget.Insets{
	//Left:   30,
	//Right:  30,
	//Top:    5,
	//Bottom: 5,
	//}),

	//widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
	//actionDispatcher.GoHome()
	//}),
	//)

	//statsBtnW := widget.NewButton(
	//widget.ButtonOpts.Image(cutils.ButtonImage),

	//widget.ButtonOpts.Text("STATS", cutils.SmallFont, &widget.ButtonTextColor{
	//Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
	//}),

	//widget.ButtonOpts.TextPadding(widget.Insets{
	//Left:   30,
	//Right:  30,
	//Top:    5,
	//Bottom: 5,
	//}),

	//widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
	//actionDispatcher.ToggleStats()
	//}),
	//)

	//topRightStatsC := widget.NewContainer(
	//widget.ContainerOpts.Layout(widget.NewRowLayout(
	//widget.RowLayoutOpts.Direction(widget.DirectionVertical),
	//widget.RowLayoutOpts.Spacing(20),
	//)),
	//widget.ContainerOpts.WidgetOpts(
	//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
	//Stretch: true,
	//}),
	//),
	//)

	//entries := make([]any, 0, 0)
	//statsListW := widget.NewList(
	//// Set the entries in the list
	//widget.ListOpts.Entries(entries),
	//widget.ListOpts.ScrollContainerOpts(
	//// Set the background images/color for the list
	//widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
	//Idle:     image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
	//Disabled: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
	//Mask:     image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
	//}),
	//),
	//widget.ListOpts.SliderOpts(
	//// Set the background images/color for the background of the slider track
	//widget.SliderOpts.Images(&widget.SliderTrackImage{
	//Idle:  image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
	//Hover: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
	//}, cutils.ButtonImage),
	//widget.SliderOpts.MinHandleSize(5),
	//// Set how wide the track should be
	//widget.SliderOpts.TrackPadding(widget.NewInsetsSimple(2)),
	//),
	//// Hide the horizontal slider
	//widget.ListOpts.HideHorizontalSlider(),
	//widget.ListOpts.HideVerticalSlider(),
	//// Set the font for the list options
	//widget.ListOpts.EntryFontFace(cutils.SmallFont),
	//// Set the colors for the list
	//widget.ListOpts.EntryColor(&widget.ListEntryColor{
	//Selected:                   color.NRGBA{254, 255, 255, 255},             // Foreground color for the unfocused selected entry
	//Unselected:                 color.NRGBA{254, 255, 255, 255},             // Foreground color for the unfocused unselected entry
	//SelectedBackground:         color.NRGBA{R: 100, G: 100, B: 100, A: 255}, // Background color for the unfocused selected entry
	//SelectedFocusedBackground:  color.NRGBA{R: 100, G: 100, B: 100, A: 255}, // Background color for the focused selected entry
	//FocusedBackground:          color.NRGBA{R: 170, G: 170, B: 180, A: 255}, // Background color for the focused unselected entry
	//DisabledUnselected:         color.NRGBA{100, 100, 100, 255},             // Foreground color for the disabled unselected entry
	//DisabledSelected:           color.NRGBA{100, 100, 100, 255},             // Foreground color for the disabled selected entry
	//DisabledSelectedBackground: color.NRGBA{100, 100, 100, 255},             // Background color for the disabled selected entry
	//}),
	//// This required function returns the string displayed in the list
	//widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
	//return e.(cutils.ListEntry).Text
	//}),
	//// Padding for each entry
	//widget.ListOpts.EntryTextPadding(widget.NewInsetsSimple(5)),
	//// Text position for each entry
	//widget.ListOpts.EntryTextPosition(widget.TextPositionStart, widget.TextPositionCenter),
	//// This handler defines what function to run when a list item is selected.
	//widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
	////entry := args.Entry.(ListEntry)
	////fmt.Println("Entry Selected: ", entry)
	//}),
	//)

	//incomeTextW := widget.NewText(
	//widget.TextOpts.Text("Gold: 40     Income Timer: 15s", cutils.SmallFont, cutils.White),
	//widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	//widget.TextOpts.WidgetOpts(
	//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
	//Position: widget.RowLayoutPositionStart,
	//}),
	//),
	//)

	//bottomRightContainer := widget.NewContainer(
	//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	//)

	//// Create the first tab
	//// A TabBookTab is a labelled container. The text here is what will show up in the tab button
	//tabUnits := widget.NewTabBookTab("UNITS",
	//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})),
	//)

	//unitsC := widget.NewContainer(
	//// the container will use an anchor layout to layout its single child widget
	//widget.ContainerOpts.Layout(widget.NewGridLayout(
	////Define number of columns in the grid
	//widget.GridLayoutOpts.Columns(5),
	////Define how much padding to inset the child content
	//widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
	////Define how far apart the rows and columns should be
	//widget.GridLayoutOpts.Spacing(5, 5),
	////Define how to stretch the rows and columns. Note it is required to
	////specify the Stretch for each row and column.
	//widget.GridLayoutOpts.Stretch([]bool{false, false, false, false, false}, []bool{false, false, false, false, false}),
	//)),
	//)
	//for _, u := range sortedUnits() {
	//uu := cp.UnitUpdates[u.Type.String()]

	//tooltipContainer := widget.NewContainer(
	//widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
	//widget.ContainerOpts.AutoDisableChildren(),
	//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
	//)

	//toolTxt := widget.NewText(
	//widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	//widget.TextOpts.Text(fmt.Sprintf(unitToolTipTmpl, uu.Level, uu.Current.Gold, uu.Current.Health, uu.Current.MovementSpeed, uu.Current.Income, u.Environment, u.Keybind), cutils.SmallFont, cutils.White),
	//widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
	//)
	//hs.unitsTooltip[u.Type.String()] = toolTxt
	//tooltipContainer.AddChild(toolTxt)

	//ubtn := widget.NewButton(
	//// set general widget options
	//widget.ButtonOpts.WidgetOpts(
	//widget.WidgetOpts.LayoutData(widget.GridLayoutData{
	//MaxWidth:  38,
	//MaxHeight: 38,
	//}),
	//widget.WidgetOpts.ToolTip(widget.NewToolTip(
	//widget.ToolTipOpts.Content(tooltipContainer),
	////widget.WidgetToolTipOpts.Delay(1*time.Second),
	//widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
	//widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
	////When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
	////They will default to what you see below if you do not provide them
	//widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
	//widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
	//widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
	//widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
	//)),
	//),

	//// specify the images to sue
	//widget.ButtonOpts.Image(cutils.ButtonImageFromImage(cutils.Images.Get(u.FacesetKey()))),

	//// add a handler that reacts to clicking the button
	//widget.ButtonOpts.ClickedHandler(func(u *unit.Unit) func(args *widget.ButtonClickedEventArgs) {
	//return func(args *widget.ButtonClickedEventArgs) {
	//cp := hs.game.Store.Lines.FindCurrentPlayer()
	//actionDispatcher.SummonUnit(u.Type.String(), cp.ID, cp.LineID, hs.game.Store.Map.GetNextLineID(cp.LineID))
	//}
	//}(u)),
	//)
	//unitsC.AddChild(ubtn)
	//}
	//hs.unitsC = unitsC
	//tabUnits.AddChild(unitsC)

	//tabTowers := widget.NewTabBookTab("TOWERS",
	//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})),
	//)
	//towersC := widget.NewContainer(
	//// the container will use an anchor layout to layout its single child widget
	//widget.ContainerOpts.Layout(widget.NewGridLayout(
	////Define number of columns in the grid
	//widget.GridLayoutOpts.Columns(1),
	////Define how much padding to inset the child content
	//widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
	////Define how far apart the rows and columns should be
	//widget.GridLayoutOpts.Spacing(5, 5),
	////Define how to stretch the rows and columns. Note it is required to
	////specify the Stretch for each row and column.
	//widget.GridLayoutOpts.Stretch([]bool{false, false, false, false, false}, []bool{false, false, false, false, false}),
	//)),
	//)
	//for _, t := range sortedTowers() {
	//tooltipContainer := widget.NewContainer(
	//widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
	//widget.ContainerOpts.AutoDisableChildren(),
	//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
	//)

	//kb := rangeTowerKeybind
	//if t.Type == tower.Melee1 {
	//kb = meleeTowerKeybind
	//}
	//toolTxt := widget.NewText(
	//widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	//widget.TextOpts.Text(fmt.Sprintf("Gold: %d\nRange: %.0f\nDamage: %.0f\nTargets: %s\nKeybind: %s", t.Gold, t.Range, t.Damage, t.Targets, kb), cutils.SmallFont, cutils.White),
	//widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
	//)
	//tooltipContainer.AddChild(toolTxt)
	//tbtn := widget.NewButton(
	//// set general widget options
	//widget.ButtonOpts.WidgetOpts(
	//widget.WidgetOpts.LayoutData(widget.GridLayoutData{
	//MaxWidth:  38,
	//MaxHeight: 38,
	//}),
	//widget.WidgetOpts.ToolTip(widget.NewToolTip(
	//widget.ToolTipOpts.Content(tooltipContainer),
	////widget.WidgetToolTipOpts.Delay(1*time.Second),
	//widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
	//widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
	////When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
	////They will default to what you see below if you do not provide them
	//widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
	//widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
	//widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
	//widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START)))), // specify the images to sue
	//widget.ButtonOpts.Image(cutils.ButtonImageFromImage(cutils.Images.Get(t.FacesetKey()))),

	//// add a handler that reacts to clicking the button
	//widget.ButtonOpts.ClickedHandler(func(t *tower.Tower) func(args *widget.ButtonClickedEventArgs) {
	//return func(args *widget.ButtonClickedEventArgs) {
	//hst := hs.GetState().(HUDState)
	//actionDispatcher.SelectTower(t.Type.String(), int(hst.LastCursorPosition.X), int(hst.LastCursorPosition.Y))
	//}
	//}(t)),
	//)
	//towersC.AddChild(tbtn)
	//}
	//hs.towersC = towersC
	//tabTowers.AddChild(towersC)

	//// Create the first tab
	//// A TabBookTab is a labelled container. The text here is what will show up in the tab button
	//tabUnitsUpdates := widget.NewTabBookTab("UNITS UPDATES",
	//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})),
	//)

	//unitUpdatesC := widget.NewContainer(
	//// the container will use an anchor layout to layout its single child widget
	//widget.ContainerOpts.Layout(widget.NewGridLayout(
	////Define number of columns in the grid
	//widget.GridLayoutOpts.Columns(5),
	////Define how much padding to inset the child content
	//widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
	////Define how far apart the rows and columns should be
	//widget.GridLayoutOpts.Spacing(5, 5),
	////Define how to stretch the rows and columns. Note it is required to
	////specify the Stretch for each row and column.
	//widget.GridLayoutOpts.Stretch([]bool{false, false, false, false, false}, []bool{false, false, false, false, false}),
	//)),
	//)
	//for _, u := range sortedUnits() {
	//uu := cp.UnitUpdates[u.Type.String()]

	//tooltipContainer := widget.NewContainer(
	//widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
	//widget.ContainerOpts.AutoDisableChildren(),
	//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
	//)

	//toolTxt := widget.NewText(
	//widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	//widget.TextOpts.Text(fmt.Sprintf(unitUpdateToolTipTmpl, uu.Level+1, uu.UpdateCost, uu.Next.Gold, uu.Next.Health, uu.Next.Income), cutils.SmallFont, cutils.White),
	//widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
	//)
	//hs.unitUpdatesTooltip[u.Type.String()] = toolTxt
	//tooltipContainer.AddChild(toolTxt)

	//ubtn := widget.NewButton(
	//// set general widget options
	//widget.ButtonOpts.WidgetOpts(
	//widget.WidgetOpts.LayoutData(widget.GridLayoutData{
	//MaxWidth:  38,
	//MaxHeight: 38,
	//}),
	//widget.WidgetOpts.ToolTip(widget.NewToolTip(
	//widget.ToolTipOpts.Content(tooltipContainer),
	////widget.WidgetToolTipOpts.Delay(1*time.Second),
	//widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
	//widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
	////When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
	////They will default to what you see below if you do not provide them
	//widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
	//widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
	//widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
	//widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
	//)),
	//),

	//// specify the images to sue
	//widget.ButtonOpts.Image(cutils.ButtonImageFromImage(cutils.Images.Get(u.FacesetKey()))),

	//// add a handler that reacts to clicking the button
	//widget.ButtonOpts.ClickedHandler(func(u *unit.Unit) func(args *widget.ButtonClickedEventArgs) {
	//return func(args *widget.ButtonClickedEventArgs) {
	//cp := hs.game.Store.Lines.FindCurrentPlayer()
	//actionDispatcher.UpdateUnit(cp.ID, u.Type.String())
	//}
	//}(u)),
	//)
	//unitUpdatesC.AddChild(ubtn)
	//}
	//hs.unitUpdatesC = unitUpdatesC
	//tabUnitsUpdates.AddChild(unitUpdatesC)

	//tabBook := widget.NewTabBook(
	//widget.TabBookOpts.TabButtonImage(cutils.ButtonImage),
	//widget.TabBookOpts.TabButtonText(cutils.SmallFont, &widget.ButtonTextColor{Idle: cutils.White, Disabled: cutils.White}),
	//widget.TabBookOpts.TabButtonSpacing(0),
	//widget.TabBookOpts.ContainerOpts(
	//widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
	//HorizontalPosition: widget.AnchorLayoutPositionEnd,
	//VerticalPosition:   widget.AnchorLayoutPositionEnd,
	//})),
	//),
	//widget.TabBookOpts.TabButtonOpts(
	//widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(5)),
	//widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(98, 0)),
	//),
	//widget.TabBookOpts.Tabs(tabUnits, tabTowers, tabUnitsUpdates),
	//)
	//bottomRightContainer.AddChild(tabBook)

	//bottomLeftContainer := hs.guiBottomLeft()
	//bottomLeftContainer.GetWidget().Visibility = widget.Visibility_Hide

	//hs.incomeTextW = incomeTextW
	//hs.statsListW = statsListW

	//topRightStatsC.AddChild(incomeTextW)
	//topRightStatsC.AddChild(statsListW)

	//topRightHorizontalRowC.AddChild(statsBtnW)
	//topRightHorizontalRowC.AddChild(homeBtnW)
	//topRightVerticalRowWraperC.AddChild(topRightHorizontalRowC)
	//topRightVerticalRowC.AddChild(topRightVerticalRowWraperC)
	//topRightVerticalRowC.AddChild(topRightStatsC)
	//topRightContainer.AddChild(topRightVerticalRowC)

	//topLeftBtnContainer := widget.NewContainer(
	//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	//)

	//leaveBtnW := widget.NewButton(
	//widget.ButtonOpts.WidgetOpts(
	//widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
	//HorizontalPosition: widget.AnchorLayoutPositionStart,
	//VerticalPosition:   widget.AnchorLayoutPositionStart,
	//}),
	//),

	//widget.ButtonOpts.Image(cutils.ButtonImage),

	//widget.ButtonOpts.Text("LEAVE", cutils.SmallFont, &widget.ButtonTextColor{
	//Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
	//}),

	//widget.ButtonOpts.TextPadding(widget.Insets{
	//Left:   30,
	//Right:  30,
	//Top:    5,
	//Bottom: 5,
	//}),

	//widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
	//u := hs.game.Store.Lines.FindCurrentPlayer()
	//actionDispatcher.RemovePlayer(u.ID)
	//}),
	//)
	//topLeftBtnContainer.AddChild(leaveBtnW)

	//centerTextContainer := widget.NewContainer(
	//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	//)

	//winLoseTextW := widget.NewText(
	//widget.TextOpts.Text("", cutils.SmallFont, cutils.White),
	//widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	//widget.TextOpts.WidgetOpts(
	//widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
	//HorizontalPosition: widget.AnchorLayoutPositionCenter,
	//VerticalPosition:   widget.AnchorLayoutPositionCenter,
	//}),
	//),
	//)
	//centerTextContainer.AddChild(winLoseTextW)
	//winLoseTextW.GetWidget().Visibility = widget.Visibility_Hide
	//hs.winLoseTextW = winLoseTextW

	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
	)

	hs.ui = &ebitenui.UI{
		Container: rootContainer,
	}

	rootContainer.AddChild(hs.displayDefaultUI())
	rootContainer.AddChild(hs.displayTargetUI())
	rootContainer.AddChild(hs.infoUI())
	rootContainer.AddChild(hs.menuUI())
	hs.scoreboardModal()
	hs.menuModal()
	hs.menuKeybindsModal()
}

func (hs *HUDStore) displayDefaultUI() *widget.Container {
	displayC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	contentRC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
		),
	)

	unitsGC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(5),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(
				widget.Insets{
					Left:   30,
					Right:  26,
					Top:    30,
					Bottom: 22,
				},
			),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(2, 2),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false, false, false, false, false}, []bool{false, false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.DisplayDefaultBGKey, 1, 1, !isPressed)),
	)

	towersGC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(
				widget.Insets{
					Left:   7,
					Right:  26,
					Top:    30,
					Bottom: 22,
				},
			),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(2, 2),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false}, []bool{false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.DisplayDefaultTowersBGKey, 1, 1, !isPressed)),
	)

	cp := hs.game.Store.Lines.FindCurrentPlayer()

	for _, u := range sortedUnits() {
		uu := cp.UnitUpdates[u.Type.String()]

		imageBtnC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewStackedLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					MaxWidth:  46,
					MaxHeight: 46,
				}),
			),
		)

		tooltipC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(20)),
				widget.RowLayoutOpts.Spacing(5),
			)),
			widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.ToolTipBGKey, 8, 8, !isPressed)),
		)

		tooltipDetailsC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				//Define number of columns in the grid
				//widget.GridLayoutOpts.Columns(2),
				widget.GridLayoutOpts.Columns(1),
				//Define how much padding to inset the child content
				//Define how far apart the rows and columns should be
				widget.GridLayoutOpts.Spacing(10, 0),
				//Define how to stretch the rows and columns. Note it is required to
				//specify the Stretch for each row and column.
				widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Stretch: true,
				}),
			),
			widget.ContainerOpts.AutoDisableChildren(),
		)

		tooltipDetailsRows := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				//Define number of columns in the grid
				//widget.GridLayoutOpts.Columns(2),
				widget.GridLayoutOpts.Columns(3),
				//Define how far apart the rows and columns should be
				widget.GridLayoutOpts.Spacing(20, 0),
				//Define how to stretch the rows and columns. Note it is required to
				//specify the Stretch for each row and column.
				widget.GridLayoutOpts.Stretch([]bool{false, false, true}, []bool{false, false}),
			)),
		)

		ttTitleTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf(unitToolTipTitleTmpl, u.Name(), u.Keybind), cutils.SmallFont, cutils.White),
		)

		goldIconC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)
		ttGoldGW := widget.NewGraphic(
			widget.GraphicOpts.Image(cutils.Images.Get(cutils.GoldIconKey)),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
		goldIconC.AddChild(ttGoldGW)

		ttGoldTxtW := widget.NewText(
			widget.TextOpts.Text(fmt.Sprint(uu.Current.Gold), cutils.SFont20, cutils.White),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)

		incomeIconC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)
		ttIncomeGW := widget.NewGraphic(
			widget.GraphicOpts.Image(cutils.Images.Get(cutils.IncomeIconKey)),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
		incomeIconC.AddChild(ttIncomeGW)

		ttIncomeTxtW := widget.NewText(
			widget.TextOpts.Text(fmt.Sprint(uu.Current.Income), cutils.SFont20, cutils.White),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)

		healthIconC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)
		ttHealthGW := widget.NewGraphic(
			widget.GraphicOpts.Image(cutils.Images.Get(cutils.HeartIconKey)),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
		healthIconC.AddChild(ttHealthGW)

		ttHealthTxtW := widget.NewText(
			widget.TextOpts.Text(fmt.Sprint(uu.Current.Health), cutils.SFont20, cutils.White),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)

		movementSpeedIconC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)
		ttMovementSpeedGW := widget.NewGraphic(
			widget.GraphicOpts.Image(cutils.Images.Get(cutils.MovementSpeedIconKey)),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
		movementSpeedIconC.AddChild(ttMovementSpeedGW)
		ttMovementSpeedTxtW := widget.NewText(
			widget.TextOpts.Text(fmt.Sprint(uu.Current.MovementSpeed), cutils.SFont20, cutils.White),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)

		nGoldC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			)),
		)
		ttNextGoldTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("Gold: %d", uu.Current.Gold), cutils.SFont20, cutils.White),
		)
		ttNextGoldDiffTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("Gold: %d", uu.Current.Gold), cutils.SFont20, cutils.Green),
		)
		nGoldC.AddChild(ttNextGoldTxt, ttNextGoldDiffTxt)

		nIncomeC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			)),
		)
		ttNextIncomeTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("Income: %d", uu.Current.Income), cutils.SFont20, cutils.White),
		)
		ttNextIncomeDiffTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("Income: %d", uu.Current.Income), cutils.SFont20, cutils.Green),
		)
		nIncomeC.AddChild(ttNextIncomeTxt, ttNextIncomeDiffTxt)

		nHealthC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			)),
		)
		ttNextHealthTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("Health: %.0f", uu.Current.Health), cutils.SFont20, cutils.White),
		)
		ttNextHealthDiffTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("Health: %.0f", uu.Current.Health), cutils.SFont20, cutils.Green),
		)
		nHealthC.AddChild(ttNextHealthTxt, ttNextHealthDiffTxt)

		nMovementSpeedC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			)),
		)
		ttNextMovementSpeedTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("Movement Speed: %.0f", uu.Current.Health), cutils.SFont20, cutils.White),
		)
		ttNextMovementSpeedDiffTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("Movement Speed: %.0f", uu.Current.Health), cutils.SFont20, cutils.Green),
		)
		nMovementSpeedC.AddChild(ttNextMovementSpeedTxt, ttNextMovementSpeedDiffTxt)

		ttCurrentLvlTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text("X", cutils.SFont20, cutils.White),
		)

		ttNextLvlTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text("Y", cutils.SFont20, cutils.White),
		)

		tooltipDetailsRows.AddChild(
			widget.NewText(
				widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
				widget.TextOpts.Text("lvl", cutils.SFont20, cutils.White),
			),
			ttCurrentLvlTxt,
			ttNextLvlTxt,

			goldIconC,
			ttGoldTxtW,
			nGoldC,

			incomeIconC,
			ttIncomeTxtW,
			nIncomeC,

			healthIconC,
			ttHealthTxtW,
			nHealthC,

			movementSpeedIconC,
			ttMovementSpeedTxtW,
			nMovementSpeedC,
		)
		hs.unitsTooltip[u.Type.String()] = &unitTooltips{
			title:      ttTitleTxt,
			currentLvl: ttCurrentLvlTxt,
			nextLvl:    ttNextLvlTxt,
			//ucost:          ttUpdateCostTxtW,
			gold:               ttGoldTxtW,
			ngold:              ttNextGoldTxt,
			ngoldDiff:          ttNextGoldDiffTxt,
			income:             ttIncomeTxtW,
			nincome:            ttNextIncomeTxt,
			nincomeDiff:        ttNextIncomeDiffTxt,
			health:             ttHealthTxtW,
			nhealth:            ttNextHealthTxt,
			nhealthDiff:        ttNextHealthDiffTxt,
			movementSpeed:      ttMovementSpeedTxtW,
			nmovementSpeed:     ttNextMovementSpeedTxt,
			nmovementSpeedDiff: ttNextMovementSpeedDiffTxt,
		}

		tooltipDetailsC.AddChild(
			tooltipDetailsRows,
		)

		ttAbilitiesTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text("Abilities:", cutils.SFont20, cutils.White),
		)

		tooltipC.AddChild(
			ttTitleTxt,
			tooltipDetailsC,
			//ttUpdateCostTxtW,
			ttAbilitiesTxt,
		)

		for _, a := range u.Abilities {
			tooltipC.AddChild(widget.NewText(
				widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
				widget.TextOpts.Text(fmt.Sprintf("%s: %s", ability.Name(a), ability.Description(a)), cutils.SFont20, cutils.White),
			))
		}

		ubtn := widget.NewButton(
			// set general widget options
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.ToolTip(widget.NewToolTip(
					widget.ToolTipOpts.Content(tooltipC),
					//widget.WidgetToolTipOpts.Delay(1*time.Second),
					widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
					widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
					//When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
					//They will default to what you see below if you do not provide them
					widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
					widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
					widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
					widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
				)),
			),

			// specify the images to sue
			widget.ButtonOpts.Image(cutils.ButtonBorderResource()),

			// add a handler that reacts to clicking the button
			widget.ButtonOpts.ClickedHandler(func(u *unit.Unit) func(args *widget.ButtonClickedEventArgs) {
				return func(args *widget.ButtonClickedEventArgs) {
					cp := hs.game.Store.Lines.FindCurrentPlayer()
					if ebiten.IsKeyPressed(ebiten.KeyShift) {
						actionDispatcher.UpdateUnit(cp.ID, u.Type.String())
					} else {
						actionDispatcher.SummonUnit(u.Type.String(), cp.ID, cp.LineID, hs.game.Store.Map.GetNextLineID(cp.LineID))
					}
				}
			}(u)),
		)

		uimg := cutils.Images.Get(u.FacesetKey())
		enabledImageBtnGraphicW := widget.NewGraphic(widget.GraphicOpts.Image(uimg))
		disabledImageBtnGraphicW := widget.NewGraphic(
			widget.GraphicOpts.Image(cutils.DisableImage(uimg)),
			widget.GraphicOpts.WidgetOpts(
				func(w *widget.Widget) {
					w.Visibility = widget.Visibility_Hide
				},
			),
		)

		keyBindC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Padding(
					widget.Insets{
						Left: 4,
						Top:  4,
					},
				),
			)),
		)
		keyBindC2 := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout(
				widget.AnchorLayoutOpts.Padding(
					widget.Insets{
						Left: 4,
						// TODO: Something from fc41b2ae37923f85ce7f562f69bc40803433eb1f to fc41b2ae37923f85ce7f562f69bc40803433eb1f broke this
						//Top:    -3,
						//Bottom: -3,
					},
				),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionStart,
				}),
			),
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(cutils.BlackT)),
		)
		keyBindTxtW := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
			widget.TextOpts.Text(u.Keybind, cutils.SmallFont, cutils.White),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionStart,
				}),
			),
		)
		keyBindC2.AddChild(keyBindTxtW)
		keyBindC.AddChild(keyBindC2)

		imageBtnC.AddChild(ubtn)
		imageBtnC.AddChild(enabledImageBtnGraphicW)
		imageBtnC.AddChild(disabledImageBtnGraphicW)
		imageBtnC.AddChild(keyBindC)

		unitUpdateAnimations := make([]*widget.Graphic, 4)
		img := cutils.Images.Get(cutils.UnitUpdateButtonAnimationKey)
		for i := 0; i < 4; i++ {
			nimg := ebiten.NewImageFromImage(img.SubImage(stdimage.Rect(i*38, 0, i*38+38, i*38+38)))
			unitUpdateAnimations[i] = widget.NewGraphic(
				widget.GraphicOpts.Image(nimg),
				widget.GraphicOpts.WidgetOpts(
					func(w *widget.Widget) {
						w.Visibility = widget.Visibility_Hide
					},
				),
			)
			imageBtnC.AddChild(unitUpdateAnimations[i])
		}

		unitsGC.AddChild(imageBtnC)

		hs.unitsBtns[u.Type.String()] = &btnWithGraphic{
			btn:      ubtn,
			enabled:  enabledImageBtnGraphicW,
			disabled: disabledImageBtnGraphicW,
		}
		hs.unitsUpdateAnimation[u.Type.String()] = unitUpdateAnimations
	}
	for _, t := range sortedTowers() {
		imageBtnC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewStackedLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					MaxWidth:  46,
					MaxHeight: 46,
				}),
			),
		)

		kb := rangeTowerKeybind
		if t.Type == tower.Melee1 {
			kb = meleeTowerKeybind
		}

		//tooltipC := widget.NewContainer(
		//widget.ContainerOpts.Layout(widget.NewRowLayout(
		//widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		//widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(20)),
		//widget.RowLayoutOpts.Spacing(5),
		//)),
		//widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.ToolTipBGKey, 8, 8, !isPressed)),
		//)

		//tooltipDetailsC := widget.NewContainer(
		//widget.ContainerOpts.Layout(widget.NewGridLayout(
		////Define number of columns in the grid
		////widget.GridLayoutOpts.Columns(2),
		//widget.GridLayoutOpts.Columns(1),
		////Define how much padding to inset the child content
		////Define how far apart the rows and columns should be
		//widget.GridLayoutOpts.Spacing(10, 0),
		////Define how to stretch the rows and columns. Note it is required to
		////specify the Stretch for each row and column.
		//widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),
		//)),
		//widget.ContainerOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
		//Stretch: true,
		//}),
		//),
		//widget.ContainerOpts.AutoDisableChildren(),
		//)

		//tooltipDetailsRows := widget.NewContainer(
		//widget.ContainerOpts.Layout(widget.NewGridLayout(
		////Define number of columns in the grid
		////widget.GridLayoutOpts.Columns(2),
		//widget.GridLayoutOpts.Columns(2),
		////Define how far apart the rows and columns should be
		//widget.GridLayoutOpts.Spacing(20, 0),
		////Define how to stretch the rows and columns. Note it is required to
		////specify the Stretch for each row and column.
		//widget.GridLayoutOpts.Stretch([]bool{false, false}, []bool{false, false}),
		//)),
		//)

		//ttTitleTxt := widget.NewText(
		//widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		//widget.TextOpts.Text(fmt.Sprintf(unitToolTipTitleTmpl, t.Name(), kb), cutils.SmallFont, cutils.White),
		//)

		//goldIconC := widget.NewContainer(
		//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		//)
		//ttGoldGW := widget.NewGraphic(
		//widget.GraphicOpts.Image(cutils.Images.Get(cutils.GoldIconKey)),
		//widget.GraphicOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
		//HorizontalPosition: widget.AnchorLayoutPositionStart,
		//VerticalPosition:   widget.AnchorLayoutPositionCenter,
		//}),
		//),
		//)
		//goldIconC.AddChild(ttGoldGW)

		//ttGoldTxtW := widget.NewText(
		//widget.TextOpts.Text(fmt.Sprint(t.Gold), cutils.SFont20, cutils.White),
		//widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		//widget.TextOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
		//Position: widget.RowLayoutPositionCenter,
		//}),
		//),
		//)

		//damageIconC := widget.NewContainer(
		//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		//)
		//ttDamageGW := widget.NewGraphic(
		//widget.GraphicOpts.Image(cutils.Images.Get(cutils.DamageIconKey)),
		//widget.GraphicOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
		//HorizontalPosition: widget.AnchorLayoutPositionStart,
		//VerticalPosition:   widget.AnchorLayoutPositionCenter,
		//}),
		//),
		//)
		//damageIconC.AddChild(ttDamageGW)

		//ttDamageTxtW := widget.NewText(
		//widget.TextOpts.Text(fmt.Sprint(t.Damage), cutils.SFont20, cutils.White),
		//widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		//widget.TextOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
		//Position: widget.RowLayoutPositionCenter,
		//}),
		//),
		//)

		//rangeIconC := widget.NewContainer(
		//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		//)
		//ttRangeGW := widget.NewGraphic(
		//widget.GraphicOpts.Image(cutils.Images.Get(cutils.RangeIconKey)),
		//widget.GraphicOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
		//HorizontalPosition: widget.AnchorLayoutPositionStart,
		//VerticalPosition:   widget.AnchorLayoutPositionCenter,
		//}),
		//),
		//)
		//rangeIconC.AddChild(ttRangeGW)

		//ttRangeTxtW := widget.NewText(
		//widget.TextOpts.Text(fmt.Sprint(t.Range), cutils.SFont20, cutils.White),
		//widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		//widget.TextOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
		//Position: widget.RowLayoutPositionCenter,
		//}),
		//),
		//)

		//healthIconC := widget.NewContainer(
		//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		//)
		//ttHealthGW := widget.NewGraphic(
		//widget.GraphicOpts.Image(cutils.Images.Get(cutils.HeartIconKey)),
		//widget.GraphicOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
		//HorizontalPosition: widget.AnchorLayoutPositionStart,
		//VerticalPosition:   widget.AnchorLayoutPositionCenter,
		//}),
		//),
		//)
		//healthIconC.AddChild(ttHealthGW)

		//ttHealthTxtW := widget.NewText(
		//widget.TextOpts.Text(fmt.Sprint(t.Health), cutils.SFont20, cutils.White),
		//widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		//widget.TextOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
		//Position: widget.RowLayoutPositionCenter,
		//}),
		//),
		//)
		//ttDescriptionTxt := widget.NewText(
		//widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		//widget.TextOpts.Text("Description:", cutils.SFont20, cutils.White),
		//)
		//ttDescriptionContentTxt := widget.NewText(
		//widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		//widget.TextOpts.Text(t.Description(), cutils.SFont20, cutils.White),
		//)

		//tooltipDetailsRows.AddChild(
		//goldIconC,
		//ttGoldTxtW,

		//rangeIconC,
		//ttRangeTxtW,

		//damageIconC,
		//ttDamageTxtW,

		//healthIconC,
		//ttHealthTxtW,
		//)

		//tooltipDetailsC.AddChild(
		//tooltipDetailsRows,
		//)

		//tooltipC.AddChild(
		//ttTitleTxt,
		//tooltipDetailsC,
		//ttDescriptionTxt,
		//ttDescriptionContentTxt,
		//)

		ttC, _, _, _, _, _, _ := hs.towerToolTip(t, kb)
		tbtn := widget.NewButton(
			// set general widget options
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
					//MaxWidth:  38,
					//MaxHeight: 38,
				}),
				widget.WidgetOpts.ToolTip(ttC),
			),

			// specify the images to sue
			widget.ButtonOpts.Image(cutils.ButtonBorderResource()),

			// add a handler that reacts to clicking the button
			widget.ButtonOpts.ClickedHandler(func(t *tower.Tower) func(args *widget.ButtonClickedEventArgs) {
				return func(args *widget.ButtonClickedEventArgs) {
					hst := hs.GetState().(HUDState)
					actionDispatcher.SelectTower(t.Type.String(), int(hst.LastCursorPosition.X), int(hst.LastCursorPosition.Y))
				}
			}(t)),
		)
		imageBtnGraphicW := widget.NewGraphic(widget.GraphicOpts.Image(cutils.Images.Get(t.FacesetKey())))

		keyBindC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Padding(
					widget.Insets{
						Left: 4,
						Top:  4,
					},
				),
			)),
		)
		keyBindC2 := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout(
				widget.AnchorLayoutOpts.Padding(
					widget.Insets{
						Left: 4,
						// TODO: Something from fc41b2ae37923f85ce7f562f69bc40803433eb1f to fc41b2ae37923f85ce7f562f69bc40803433eb1f broke this
						//Top:    -3,
						//Bottom: -3,
					},
				),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionStart,
				}),
			),
			//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(cutils.BlackT)),
		)
		keyBindTxtW := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
			widget.TextOpts.Text(kb.String(), cutils.SmallFont, cutils.White),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionStart,
				}),
			),
		)
		keyBindC2.AddChild(keyBindTxtW)
		keyBindC.AddChild(keyBindC2)

		imageBtnC.AddChild(
			tbtn,
			imageBtnGraphicW,
			keyBindC,
		)

		towersGC.AddChild(imageBtnC)
	}

	contentRC.AddChild(unitsGC)
	contentRC.AddChild(towersGC)

	displayC.AddChild(contentRC)

	hs.displayDefaultC = displayC

	return displayC
}

func (hs *HUDStore) displayTargetUI() *widget.Container {
	displayC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			func(w *widget.Widget) {
				w.Visibility = widget.Visibility_Hide
			},
		),
	)

	contentRC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
		),
	)

	imageC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(
				widget.Insets{
					Left:   24,
					Right:  16,
					Top:    24,
					Bottom: 16,
				},
			),
		)),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.DisplayTargetImageBGKey, 1, 1, !isPressed)),

		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(146, 146),
		),
	)
	imageG := widget.NewGraphic(
		// TODO: Image 99x99 the image
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)
	healthPB := widget.NewProgressBar(
		widget.ProgressBarOpts.Direction(widget.DirectionHorizontal),
		widget.ProgressBarOpts.Images(
			// Track
			&widget.ProgressBarImage{
				Idle: image.NewNineSliceColor(color.NRGBA{R: 95, G: 113, B: 96, A: 255}),
			},
			// Fill
			&widget.ProgressBarImage{
				Idle: image.NewNineSliceColor(color.NRGBA{R: 224, G: 57, B: 76, A: 255}),
			},
		),
		widget.ProgressBarOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				StretchHorizontal:  true,
			}),
			widget.WidgetOpts.MinSize(0, 20),
		),
	)
	healthTxt := widget.NewText(
		widget.TextOpts.Text("20/20", cutils.SFont20, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionEnd),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				//StretchHorizontal:  true,
			}),
		),
	)
	imageC.AddChild(
		imageG,
		healthPB,
		healthTxt,
	)

	// NOTE: Now we have to have 2 types of container, one for
	// towers and one for units as they have different info

	towerDetailsC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(2),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(
				widget.Insets{
					Left:   7,
					Right:  26,
					Top:    30,
					Bottom: 22,
				},
			),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(10, 0),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false, false}, []bool{false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
			widget.WidgetOpts.MinSize(229, 0),
		),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.DisplayTargetDetailsBGKey, 1, 1, !isPressed)),
	)

	towerInfoC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(2),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	towerNameTxtW := widget.NewText(
		widget.TextOpts.Text("TOWER NAME", cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		//widget.TextOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
		//MaxHeight: 10,
		//}),
		//),
	)
	towerDetailsInfoC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			//widget.RowLayoutOpts.Spacing(1),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	towerInfoRangeC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
		//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{255, 0, 255, 255})),
	)
	towerRangeGW := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.RangeIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)
	towerRangeTxtW := widget.NewText(
		widget.TextOpts.Text("20", cutils.SFont20, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.Insets(
			widget.Insets{
				//Top: -2,
				//Left:   1,
				//Right:  1,
				//Bottom: 20,
				//Bottom: 1,
			},
		),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)
	towerInfoRangeC.AddChild(
		towerRangeGW,
		towerRangeTxtW,
	)

	towerInfoDamageC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)
	towerDamageGW := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.DamageIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)
	towerDamageTxtW := widget.NewText(
		widget.TextOpts.Text("30", cutils.SFont20, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
		//MaxHeight: 15,
		//}),
		),
	)
	towerInfoDamageC.AddChild(
		towerDamageGW,
		towerDamageTxtW,
	)

	towerInfoAttackSpeedC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)
	towerAttackSpeedGW := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.AttackSpeedIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)
	towerAttackSpeedTxtW := widget.NewText(
		widget.TextOpts.Text("2", cutils.SFont20, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
		//MaxHeight: 15,
		//}),
		),
	)
	towerInfoAttackSpeedC.AddChild(
		towerAttackSpeedGW,
		towerAttackSpeedTxtW,
	)

	towerDetailsInfoC.AddChild(
		towerInfoDamageC,
		towerInfoRangeC,
		towerInfoAttackSpeedC,
	)
	towerInfoC.AddChild(
		towerNameTxtW,
		towerDetailsInfoC,
	)

	towerButtonsC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(2),
			//Define how much padding to inset the child content
			//widget.GridLayoutOpts.Padding(
			//widget.Insets{
			//Left:   7,
			//Right:  26,
			//Top:    30,
			//Bottom: 22,
			//},
			//),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(2, 2),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false, false}, []bool{false, false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
		// TODO: Add background
		//widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.DisplayDefaultTowersBGKey, 1, 1, !isPressed)),
	)

	sellImageBtnC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:  46,
				MaxHeight: 46,
			}),
		),
	)

	ttSellC, _, ttSellDescriptionTxtW := hs.simpleTooltip(fmt.Sprintf("Sell (%s)", sellTowerKeybind.String()), "Sell the tower for X amount")
	sellBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.ToolTip(ttSellC),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.ButtonBorderResource()),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
		}),
	)

	hs.displayTargetTowerSellToolTip = ttSellDescriptionTxtW

	sellImageBtnGraphicW := widget.NewGraphic(widget.GraphicOpts.Image(cutils.Images.Get(cutils.SellIconKey)))

	sellKeyBindC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(
				widget.Insets{
					Left: 4,
					Top:  4,
				},
			),
		)),
	)
	sellKeyBindC2 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(
				widget.Insets{
					Left: 4,
					// TODO: Something from fc41b2ae37923f85ce7f562f69bc40803433eb1f to fc41b2ae37923f85ce7f562f69bc40803433eb1f broke this
					//Top:    -3,
					//Bottom: -3,
				},
			),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(cutils.BlackT)),
	)
	sellKeyBindTxtW := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.Text(sellTowerKeybind.String(), cutils.SmallFont, cutils.White),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)
	sellKeyBindC2.AddChild(sellKeyBindTxtW)
	sellKeyBindC.AddChild(sellKeyBindC2)

	sellImageBtnC.AddChild(
		sellBtnW,
		sellImageBtnGraphicW,
		sellKeyBindC,
	)

	update1ImageBtnC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:  46,
				MaxHeight: 46,
			}),
		),
	)

	ttu1C, ttu1TitleTxtW, ttu1GoldTxtW, ttu1DamageTxtW, ttu1RangeTxtW, ttu1HealthTxtW, ttu1DescriptionContentTxtW := hs.towerToolTip(tower.Towers[tower.Range1.String()], updateTowerKeybind1)

	update1BtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.ToolTip(ttu1C),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.ButtonBorderResource()),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
		}),
	)

	update1ImageBtnGraphicW := widget.NewGraphic(widget.GraphicOpts.Image(cutils.Images.Get(cutils.SellIconKey)))

	update1KeyBindC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(
				widget.Insets{
					Left: 4,
					Top:  4,
				},
			),
		)),
	)
	update1KeyBindC2 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(
				widget.Insets{
					Left: 4,
					// TODO: Something from fc41b2ae37923f85ce7f562f69bc40803433eb1f to fc41b2ae37923f85ce7f562f69bc40803433eb1f broke this
					//Top:    -3,
					//Bottom: -3,
				},
			),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(cutils.BlackT)),
	)
	update1KeyBindTxtW := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.Text(updateTowerKeybind1.String(), cutils.SmallFont, cutils.White),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)

	update1KeyBindC2.AddChild(update1KeyBindTxtW)
	update1KeyBindC.AddChild(update1KeyBindC2)

	update1ImageBtnC.AddChild(
		update1BtnW,
		update1ImageBtnGraphicW,
		update1KeyBindC,
	)

	update2ImageBtnC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:  46,
				MaxHeight: 46,
			}),
		),
	)

	ttu2C, ttu2TitleTxtW, ttu2GoldTxtW, ttu2DamageTxtW, ttu2RangeTxtW, ttu2HealthTxtW, ttu2DescriptionContentTxtW := hs.towerToolTip(tower.Towers[tower.Range1.String()], updateTowerKeybind2)

	update2BtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.ToolTip(ttu2C),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.ButtonBorderResource()),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
		}),
	)

	update2ImageBtnGraphicW := widget.NewGraphic(widget.GraphicOpts.Image(cutils.Images.Get(cutils.SellIconKey)))

	update2KeyBindC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(
				widget.Insets{
					Left: 4,
					Top:  4,
				},
			),
		)),
	)
	update2KeyBindC2 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(
				widget.Insets{
					Left: 4,
					// TODO: Something from fc41b2ae37923f85ce7f562f69bc40803433eb1f to fc41b2ae37923f85ce7f562f69bc40803433eb1f broke this
					//Top:    -3,
					//Bottom: -3,
				},
			),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(cutils.BlackT)),
	)
	update2KeyBindTxtW := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.Text(updateTowerKeybind2.String(), cutils.SmallFont, cutils.White),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)
	update2KeyBindC2.AddChild(update2KeyBindTxtW)
	update2KeyBindC.AddChild(update2KeyBindC2)

	update2ImageBtnC.AddChild(
		update2BtnW,
		update2ImageBtnGraphicW,
		update2KeyBindC,
	)

	towerButtonsC.AddChild(
		// We add a first empty box on the top left
		widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		),
		sellImageBtnC,
		update1ImageBtnC,
		update2ImageBtnC,
	)

	towerDetailsC.AddChild(
		towerInfoC,
		towerButtonsC,
	)

	unitDetailsC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(2),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(
				widget.Insets{
					Left:   7,
					Right:  26,
					Top:    30,
					Bottom: 20,
				},
			),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(10, 0),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false, false}, []bool{false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
			widget.WidgetOpts.MinSize(229, 0),
		),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.DisplayTargetDetailsBGKey, 1, 1, !isPressed)),
	)

	unitInfoC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			//widget.RowLayoutOpts.Spacing(2),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	unitNameTxtW := widget.NewText(
		widget.TextOpts.Text("UNIT NAME", cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
	)

	// TODO: MS, Bounty, Player
	unitDetailsInfoC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(2),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	unitInfoMovementSpeedC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)
	unitMovementSpeedGW := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.MovementSpeedIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	unitMovementSpeedTxtW := widget.NewText(
		widget.TextOpts.Text("2", cutils.SFont20, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)
	unitInfoMovementSpeedC.AddChild(
		unitMovementSpeedGW,
		unitMovementSpeedTxtW,
	)

	unitInfoBountyC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)
	unitBountyGW := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.GoldIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	unitBountyTxtW := widget.NewText(
		widget.TextOpts.Text("2", cutils.SFont20, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)
	unitInfoBountyC.AddChild(
		unitBountyGW,
		unitBountyTxtW,
	)

	unitInfoPlayerC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)
	unitPlayerGW := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.PlayerIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	unitPlayerTxtW := widget.NewText(
		widget.TextOpts.Text("Potato", cutils.SFont20, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)
	unitInfoPlayerC.AddChild(
		unitPlayerGW,
		unitPlayerTxtW,
	)

	unitDetailsInfoC.AddChild(
		unitInfoMovementSpeedC,
		unitInfoBountyC,
		unitInfoPlayerC,
	)

	unitAbilitiesC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(2),
			//Define how much padding to inset the child content
			//widget.GridLayoutOpts.Padding(
			//widget.Insets{
			//Left:   7,
			//Right:  26,
			//Top:    30,
			//Bottom: 22,
			//},
			//),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(2, 2),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false, false}, []bool{false, false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	ab1ImageBtnC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:  46,
				MaxHeight: 46,
			}),
		),
	)

	ab1TooltipContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
		widget.ContainerOpts.AutoDisableChildren(),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
	)

	ab1ToolTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text("SELL", cutils.SmallFont, cutils.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
		),
	)
	//hs.unitsTooltip[u.Type.String()] = toolTxt
	ab1TooltipContainer.AddChild(ab1ToolTxt)

	ab1BtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.ToolTip(widget.NewToolTip(
				widget.ToolTipOpts.Content(ab1TooltipContainer),
				//widget.WidgetToolTipOpts.Delay(1*time.Second),
				widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
				widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
				//When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
				//They will default to what you see below if you do not provide them
				widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
				widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
				widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
				widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
			)),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.ButtonBorderResource()),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
		}),
	)

	ab1ImageBtnGraphicW := widget.NewGraphic(widget.GraphicOpts.Image(cutils.Images.Get(cutils.SellIconKey)))

	ab1ImageBtnC.AddChild(
		ab1BtnW,
		ab1ImageBtnGraphicW,
	)

	ab2ImageBtnC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:  46,
				MaxHeight: 46,
			}),
		),
	)

	ab2TooltipContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
		widget.ContainerOpts.AutoDisableChildren(),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
	)

	ab2ToolTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text("SELL", cutils.SmallFont, cutils.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
		),
	)
	//hs.unitsTooltip[u.Type.String()] = toolTxt
	ab2TooltipContainer.AddChild(ab2ToolTxt)

	ab2BtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.ToolTip(widget.NewToolTip(
				widget.ToolTipOpts.Content(ab2TooltipContainer),
				//widget.WidgetToolTipOpts.Delay(1*time.Second),
				widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
				widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
				//When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
				//They will default to what you see below if you do not provide them
				widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
				widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
				widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
				widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
			)),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.ButtonBorderResource()),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
		}),
	)

	ab2ImageBtnGraphicW := widget.NewGraphic(widget.GraphicOpts.Image(cutils.Images.Get(cutils.SellIconKey)))

	ab2ImageBtnC.AddChild(
		ab2BtnW,
		ab2ImageBtnGraphicW,
	)

	unitInfoC.AddChild(
		unitNameTxtW,
		unitDetailsInfoC,
	)

	unitAbilitiesC.AddChild(
		// We add a first empty box on the top left
		widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewStackedLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(46, 46),
			),
		),
		ab1ImageBtnC,
		widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewStackedLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(46, 46),
			),
		),
		widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewStackedLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(46, 46),
			),
		),
	)

	unitDetailsC.AddChild(
		unitInfoC,
		unitAbilitiesC,
	)

	contentRC.AddChild(
		imageC,
		towerDetailsC,
		unitDetailsC,
	)

	displayC.AddChild(contentRC)

	hs.displayTargetProfile = imageG
	hs.displayTargetHealth = healthTxt
	hs.displayTargetHealthBar = healthPB

	// Towers
	hs.displayTargetTowerC = towerDetailsC

	hs.displayTargetTowerUpdateC1 = update1ImageBtnC
	hs.displayTargetTowerUpdateButton1 = update1BtnW
	hs.displayTargetTowerUpdateImage1 = update1ImageBtnGraphicW
	hs.displayTargetTowerUpdateToolTip1TitleTxt = ttu1TitleTxtW
	hs.displayTargetTowerUpdateToolTip1GoldTxt = ttu1GoldTxtW
	hs.displayTargetTowerUpdateToolTip1DamageTxt = ttu1DamageTxtW
	hs.displayTargetTowerUpdateToolTip1RangeTxt = ttu1RangeTxtW
	hs.displayTargetTowerUpdateToolTip1HealthTxt = ttu1HealthTxtW
	hs.displayTargetTowerUpdateToolTip1DescriptionTxt = ttu1DescriptionContentTxtW

	hs.displayTargetTowerUpdateC2 = update2ImageBtnC
	hs.displayTargetTowerUpdateButton2 = update2BtnW
	hs.displayTargetTowerUpdateImage2 = update2ImageBtnGraphicW
	hs.displayTargetTowerUpdateToolTip2TitleTxt = ttu2TitleTxtW
	hs.displayTargetTowerUpdateToolTip2GoldTxt = ttu2GoldTxtW
	hs.displayTargetTowerUpdateToolTip2DamageTxt = ttu2DamageTxtW
	hs.displayTargetTowerUpdateToolTip2RangeTxt = ttu2RangeTxtW
	hs.displayTargetTowerUpdateToolTip2HealthTxt = ttu2HealthTxtW
	hs.displayTargetTowerUpdateToolTip2DescriptionTxt = ttu2DescriptionContentTxtW

	hs.displayTargetTowerRangeTxtW = towerRangeTxtW
	hs.displayTargetTowerDamageTxtW = towerDamageTxtW
	hs.displayTargetTowerAttackSpeedTxtW = towerAttackSpeedTxtW
	hs.displayTargetTowerNameTxtW = towerNameTxtW

	// Units
	hs.displayTargetUnitC = unitDetailsC

	hs.displayTargetUnitNameTxtW = unitNameTxtW
	hs.displayTargetUnitMovementSpeedTxtW = unitMovementSpeedTxtW
	hs.displayTargetUnitBountyTxtW = unitBountyTxtW
	hs.displayTargetUnitPlayerTxtW = unitPlayerTxtW

	hs.displayTargetUnitAbilityImage1 = ab1ImageBtnGraphicW

	hs.displayTargetC = displayC

	return displayC
}

func (hs *HUDStore) scoreboardModal() {
	frameC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(10)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(0, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),
		)),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.Border4Key, 4, 6, !isPressed)),
	)

	tableHeaderC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(6),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 0),
			// TODO: Does not work
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false, true, true, true, true, true}, []bool{false}),
			//widget.GridLayoutOpts.Stretch([]bool{false}, []bool{false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	tableHeaderC.AddChild(
		// First header is empty as it's the image
		widget.NewText(
			widget.TextOpts.Text("", cutils.NormalFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
				}),
				func(w *widget.Widget) {
					w.MinWidth = 46
				},
			),
		),
		widget.NewText(
			widget.TextOpts.Text("Name", cutils.NormalFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
					MaxWidth:           100,
				}),
				func(w *widget.Widget) {
					w.MinWidth = 100
				},
			),
		),
		widget.NewText(
			widget.TextOpts.Text("Units", cutils.NormalFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
					MaxWidth:           100,
				}),
				func(w *widget.Widget) {
					w.MinWidth = 100
				},
			),
		),
		widget.NewText(
			widget.TextOpts.Text("Lives", cutils.NormalFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
					MaxWidth:           100,
				}),
				func(w *widget.Widget) {
					w.MinWidth = 100
				},
			),
		),
		widget.NewText(
			widget.TextOpts.Text("Income", cutils.NormalFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
					MaxWidth:           100,
				}),
				func(w *widget.Widget) {
					w.MinWidth = 100
				},
			),
		),
		widget.NewText(
			widget.TextOpts.Text("Researches", cutils.NormalFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
					MaxWidth:           200,
				}),
				func(w *widget.Widget) {
					w.MinWidth = 200
				},
			),
		),
	)

	tableBodyC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(0, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	frameC.AddChild(
		tableHeaderC,
		tableBodyC,
	)

	window := widget.NewWindow(
		widget.WindowOpts.Contents(frameC),
		widget.WindowOpts.Modal(),
	)

	hs.scoreboardC = frameC
	hs.scoreboardBodyC = tableBodyC
	hs.scoreboardW = window
}

func (hs *HUDStore) buildScoreboard() {
	hs.scoreboardBodyC.RemoveChildren()

	// This are the Headers
	var sortedPlayers = make([]*store.Player, 0, 0)
	livesMax := 0
	livesMin := 0
	incomeMax := 0
	incomeMin := 0
	for _, p := range hs.game.Store.Lines.ListPlayers() {
		if p.Lives > livesMax {
			livesMax = p.Lives
		}
		if livesMin == 0 || p.Lives < livesMin && p.Lives != 0 {
			livesMin = p.Lives
		}
		if p.Income > incomeMax {
			incomeMax = p.Income
		}
		if incomeMin == 0 || p.Income < incomeMin && p.Income != 0 {
			incomeMin = p.Income
		}
		sortedPlayers = append(sortedPlayers, p)
	}
	sort.Slice(sortedPlayers, func(i, j int) bool {
		ii := sortedPlayers[i]
		jj := sortedPlayers[j]
		return ii.LineID < jj.LineID
	})
	for _, p := range sortedPlayers {
		livesColor := cutils.TextColor
		incomeColor := cutils.TextColor
		if p.Lives == livesMax {
			livesColor = cutils.Green
		} else if p.Lives == livesMin {
			livesColor = cutils.Red
		}
		if p.Income == incomeMax {
			incomeColor = cutils.Green
		} else if p.Income == incomeMin {
			incomeColor = cutils.Red
		}
		var bgk string
		if p.Current {
			bgk = cutils.ScoreboardRowCurrentBGKey
		} else {
			bgk = cutils.ScoreboardRowBGKey
		}
		tableRowC := widget.NewContainer(
			// the container will use an anchor layout to layout its single child widget
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				//Define number of columns in the grid
				widget.GridLayoutOpts.Columns(6),
				//Define how far apart the rows and columns should be
				widget.GridLayoutOpts.Spacing(20, 0),
				widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
				//Define how to stretch the rows and columns. Note it is required to
				//specify the Stretch for each row and column.
				widget.GridLayoutOpts.Stretch([]bool{false, true, true, true, true, true}, []bool{false}),
				//widget.GridLayoutOpts.Stretch([]bool{false}, []bool{false}),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
				}),
			),
			widget.ContainerOpts.BackgroundImage(cutils.ImageToNineSlice(bgk)),
		)
		// Image
		imageBtnC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewStackedLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionCenter,
					VerticalPosition:   widget.GridLayoutPositionCenter,
					MaxWidth:           46,
					MaxHeight:          46,
				}),
			),
		)

		imageBtnBtnW := widget.NewButton(
			widget.ButtonOpts.Image(cutils.ButtonImageFromKey(cutils.Border4Key, 4, 6)),
		)
		imageBtnGraphicW := widget.NewGraphic(widget.GraphicOpts.Image(cutils.Images.Get(unit.Units[unit.Ninja.String()].FacesetKey())))

		imageBtnC.AddChild(imageBtnBtnW)
		imageBtnC.AddChild(imageBtnGraphicW)
		utt, _, _ := hs.simpleTooltip("Units", "The level of the units in the order they are on the display")
		unitsDetailsC := widget.NewContainer(
			// the container will use an anchor layout to layout its single child widget
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				//Define number of columns in the grid
				widget.GridLayoutOpts.Columns(5),
				//Define how far apart the rows and columns should be
				//widget.GridLayoutOpts.Spacing(2, 0),
				//widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
				//Define how to stretch the rows and columns. Note it is required to
				//specify the Stretch for each row and column.
				widget.GridLayoutOpts.Stretch([]bool{true, true, true, true, true}, []bool{false}),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
					MaxWidth:           100,
				}),
				widget.WidgetOpts.ToolTip(utt),
				func(w *widget.Widget) {
					w.MinWidth = 100
				},
			),
		)
		for _, u := range unit.TypeStrings() {
			unitsDetailsC.AddChild(
				widget.NewText(
					widget.TextOpts.Text(fmt.Sprint(p.UnitUpdates[u].Level), cutils.NormalFont, cutils.TextColor),
					widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
				),
			)
		}
		tableRowC.AddChild(
			imageBtnC,
			widget.NewText(
				widget.TextOpts.Text(p.Name, cutils.NormalFont, cutils.TextColor),
				widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
				widget.TextOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(widget.GridLayoutData{
						HorizontalPosition: widget.GridLayoutPositionStart,
						VerticalPosition:   widget.GridLayoutPositionStart,
						MaxWidth:           100,
					}),
					func(w *widget.Widget) {
						w.MinWidth = 100
					},
				),
			),
			unitsDetailsC,
			widget.NewText(
				widget.TextOpts.Text(fmt.Sprint(p.Lives), cutils.NormalFont, livesColor),
				widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
				widget.TextOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(widget.GridLayoutData{
						HorizontalPosition: widget.GridLayoutPositionStart,
						VerticalPosition:   widget.GridLayoutPositionStart,
						MaxWidth:           100,
					}),
					func(w *widget.Widget) {
						w.MinWidth = 100
					},
				),
			),
			widget.NewText(
				widget.TextOpts.Text(fmt.Sprint(p.Income), cutils.NormalFont, incomeColor),
				widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
				widget.TextOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(widget.GridLayoutData{
						HorizontalPosition: widget.GridLayoutPositionStart,
						VerticalPosition:   widget.GridLayoutPositionStart,
						MaxWidth:           100,
					}),
					func(w *widget.Widget) {
						w.MinWidth = 100
					},
				),
			),
			widget.NewText(
				widget.TextOpts.Text("WIP", cutils.NormalFont, cutils.TextColor),
				widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
				widget.TextOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(widget.GridLayoutData{
						HorizontalPosition: widget.GridLayoutPositionStart,
						VerticalPosition:   widget.GridLayoutPositionStart,
						MaxWidth:           300,
					}),
					func(w *widget.Widget) {
						w.MinWidth = 300
					},
				),
			),
		)
		hs.scoreboardBodyC.AddChild(tableRowC)
	}
}

func (hs *HUDStore) infoUI() *widget.Container {
	infoC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	contentRC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)

	timerC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(6)),
		)),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.InfoLeftBGKey, 1, 1, !isPressed)),
	)
	timerTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text("000:01", cutils.SmallFont, cutils.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(96, 0)),
	)

	timerC.AddChild(timerTxt)

	goldInfoC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(6)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.InfoMiddleBGKey, 1, 1, !isPressed)),
	)

	goldIconG := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.GoldIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(16, 0),
		),
	)
	goldTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text("40", cutils.SmallFont, cutils.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 0)),
	)

	goldInfoC.AddChild(
		goldIconG,
		goldTxt,
	)

	capInfoC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(6)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.InfoMiddleBGKey, 1, 1, !isPressed)),
	)

	capIconG := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.CapIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(16, 0),
		),
	)
	capTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text("0/200", cutils.SmallFont, cutils.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 0)),
	)

	capInfoC.AddChild(
		capIconG,
		capTxt,
	)

	livesInfoC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(6)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),

		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.InfoMiddleBGKey, 1, 1, !isPressed)),
	)

	livesIconG := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.HeartIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(16, 0),
		),
	)
	livesTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text("20", cutils.SmallFont, cutils.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 0)),
	)

	livesInfoC.AddChild(
		livesIconG,
		livesTxt,
	)

	incomeInfoC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(6)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.InfoMiddleBGKey, 1, 1, !isPressed)),
	)

	incomeIconG := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.IncomeIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(16, 0),
		),
	)
	incomeTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text("25", cutils.SmallFont, cutils.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 0)),
	)

	incomeInfoC.AddChild(
		incomeIconG,
		incomeTxt,
	)

	incomeTimerInfoC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(6)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),

		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.InfoRightBGKey, 1, 1, !isPressed)),
	)

	incomeTimerIconG := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.IncomeTimerIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(16, 0),
		),
	)
	incomeTimerTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text("15s", cutils.SmallFont, cutils.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 0)),
	)

	incomeTimerInfoC.AddChild(
		incomeTimerIconG,
		incomeTimerTxt,
	)

	contentRC.AddChild(
		timerC,
		goldInfoC,
		capInfoC,
		livesInfoC,
		incomeInfoC,
		incomeTimerInfoC,
	)

	infoC.AddChild(contentRC)

	hs.infoTimerTxt = timerTxt
	hs.infoGoldTxt = goldTxt
	hs.infoCapTxt = capTxt
	hs.infoLivesTxt = livesTxt
	hs.infoIncomeTxt = incomeTxt
	hs.infoIncomeTimerTxt = incomeTimerTxt

	return infoC
}

func (hs *HUDStore) menuUI() *widget.Container {
	menuC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	menuBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Menu", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(
			widget.Insets{
				Left:   30,
				Right:  30,
				Top:    5,
				Bottom: 5,
			},
		),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			hs.loadModal(hs.menuW)
		}),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
	)

	hs.menuBtnW = menuBtnW
	menuC.AddChild(menuBtnW)

	return menuC
}

func (hs *HUDStore) menuModal() {
	frameC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(80)),
			widget.RowLayoutOpts.Spacing(40),
		)),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.SetupGameFrameKey, 1, 1, !isPressed)),
	)

	titleW := widget.NewText(
		widget.TextOpts.Text("Menu", cutils.Font60, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	buttonsC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(0, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true, true}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)

	btnPadding := widget.Insets{
		Left:   30,
		Right:  30,
		Top:    15,
		Bottom: 15,
	}

	keybinds := widget.NewButton(

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Keybinds", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(btnPadding),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			hs.closeModal(hs.menuW)
			hs.loadModal(hs.keybindsW)
		}),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
	)

	leave := widget.NewButton(

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Leave", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(btnPadding),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			u := hs.game.Store.Lines.FindCurrentPlayer()
			actionDispatcher.RemovePlayer(u.ID)
		}),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
	)

	buttonsC.AddChild(
		keybinds,
		leave,
	)

	frameC.AddChild(
		titleW,
		buttonsC,
	)

	window := widget.NewWindow(
		widget.WindowOpts.Contents(frameC),
		widget.WindowOpts.Modal(),
		widget.WindowOpts.CloseMode(widget.CLICK_OUT),
	)

	hs.menuW = window
}

func (hs *HUDStore) menuKeybindsModal() {
	frameC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(80)),
			widget.RowLayoutOpts.Spacing(40),
		)),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.SetupGameFrameKey, 1, 1, !isPressed)),
	)

	titleW := widget.NewText(
		widget.TextOpts.Text("Keybinds", cutils.Font40, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)

	keybindsGC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(0, 20),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true, true, true, true}, []bool{false, false, false, false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)

	unitstitleW := widget.NewText(
		widget.TextOpts.Text("Units", cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	unitsKeybindsGC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(2),
			//Define how much padding to inset the child content
			//widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(5, 5),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true, true, true, true, true, true, true, true, true, true, true}, []bool{false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(400, 0),
		),
	)

	for _, u := range sortedUnits() {
		nameTxt := widget.NewText(
			widget.TextOpts.Text(u.Name(), cutils.SmallFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		)
		keybindTxt := widget.NewText(
			widget.TextOpts.Text(u.Keybind, cutils.SmallFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		)
		unitsKeybindsGC.AddChild(
			nameTxt,
			keybindTxt,
		)
	}

	updateNameTxt := widget.NewText(
		widget.TextOpts.Text("Mod + unit keybind", cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
	updateKeybindTxt := widget.NewText(
		widget.TextOpts.Text(ebiten.KeyShift.String(), cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
	)
	unitsKeybindsGC.AddChild(
		updateNameTxt,
		updateKeybindTxt,
	)

	towerstitleW := widget.NewText(
		widget.TextOpts.Text("Towers", cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	towersKeybindsGC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(2),
			//Define how much padding to inset the child content
			//widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(5, 5),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true, true, true, true}, []bool{false, false}),
		)),
	)

	for _, t := range sortedTowers() {
		nameTxt := widget.NewText(
			widget.TextOpts.Text(t.Name(), cutils.SmallFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		)
		keybindTxt := widget.NewText(
			widget.TextOpts.Text(towerKeybinds[t.Type.String()].String(), cutils.SmallFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		)
		towersKeybindsGC.AddChild(
			nameTxt,
			keybindTxt,
		)
	}

	sellNameTxt := widget.NewText(
		widget.TextOpts.Text("Target - Sell", cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
	sellKeybindTxt := widget.NewText(
		widget.TextOpts.Text(sellTowerKeybind.String(), cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
	)
	towersKeybindsGC.AddChild(
		sellNameTxt,
		sellKeybindTxt,
	)

	firstUpdateNameTxt := widget.NewText(
		widget.TextOpts.Text("Target - First update", cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
	firstUpdateKeybindTxt := widget.NewText(
		widget.TextOpts.Text(ebiten.KeyZ.String(), cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
	)
	towersKeybindsGC.AddChild(
		firstUpdateNameTxt,
		firstUpdateKeybindTxt,
	)

	secondUpdateNameTxt := widget.NewText(
		widget.TextOpts.Text("Target - Second update", cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
	secondUpdateKeybindTxt := widget.NewText(
		widget.TextOpts.Text(ebiten.KeyX.String(), cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
	)
	towersKeybindsGC.AddChild(
		secondUpdateNameTxt,
		secondUpdateKeybindTxt,
	)

	otherstitleW := widget.NewText(
		widget.TextOpts.Text("Others", cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	othersKeybindsGC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(2),
			//Define how much padding to inset the child content
			//widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(5, 5),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{false, false}),
		)),
	)

	homeTxt := widget.NewText(
		widget.TextOpts.Text("Home Line", cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
	homeKeybindTxt := widget.NewText(
		widget.TextOpts.Text(ebiten.KeyF1.String(), cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
	)
	scoreboardTxt := widget.NewText(
		widget.TextOpts.Text("Scoreboard", cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
	scoreboardKeybindTxt := widget.NewText(
		widget.TextOpts.Text(ebiten.KeyTab.String(), cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
	)
	menuTxt := widget.NewText(
		widget.TextOpts.Text("Open/Close Menu", cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
	menuKeybindTxt := widget.NewText(
		widget.TextOpts.Text(ebiten.KeyEscape.String(), cutils.SmallFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
	)

	othersKeybindsGC.AddChild(
		homeTxt, homeKeybindTxt,
		scoreboardTxt, scoreboardKeybindTxt,
		menuTxt, menuKeybindTxt,
	)

	keybindsGC.AddChild(
		unitstitleW,
		unitsKeybindsGC,
		towerstitleW,
		towersKeybindsGC,
		otherstitleW,
		othersKeybindsGC,
	)

	frameC.AddChild(
		titleW,
		keybindsGC,
	)

	window := widget.NewWindow(
		widget.WindowOpts.Contents(frameC),
		widget.WindowOpts.Modal(),
		widget.WindowOpts.CloseMode(widget.CLICK_OUT),
	)

	hs.keybindsW = window
}

//func (hs *HUDStore) guiBottomLeft() *widget.Container {
//bottomLeftContainer := widget.NewContainer(
//widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
//)
//towerMenuC := widget.NewContainer(
//// the container will use an anchor layout to layout its single child widget
//widget.ContainerOpts.Layout(widget.NewGridLayout(
////Define number of columns in the grid
//widget.GridLayoutOpts.Columns(5),
////Define how much padding to inset the child content
//widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
////Define how far apart the rows and columns should be
//widget.GridLayoutOpts.Spacing(5, 5),
////Define how to stretch the rows and columns. Note it is required to
////specify the Stretch for each row and column.
//widget.GridLayoutOpts.Stretch([]bool{false, false, false, false, false}, []bool{false, false, false, false, false}),
//)),
//widget.ContainerOpts.WidgetOpts(
//widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
//HorizontalPosition: widget.AnchorLayoutPositionStart,
//VerticalPosition:   widget.AnchorLayoutPositionEnd,
//}),
//),
//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
//)
//// Remove button
//removeToolTipContainer := widget.NewContainer(
//widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
//widget.ContainerOpts.AutoDisableChildren(),
//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
//)

//removeToolTxt := widget.NewText(
//widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
//widget.TextOpts.Text(fmt.Sprintf(towerRemoveToolTipTmpl, 0, sellTowerKeybind), cutils.SmallFont, cutils.White),
//widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
//)
//hs.towerRemoveToolTip = removeToolTxt
//removeToolTipContainer.AddChild(removeToolTxt)

//rbtn := widget.NewButton(
//// set general widget options
//widget.ButtonOpts.WidgetOpts(
//widget.WidgetOpts.LayoutData(widget.GridLayoutData{
//MaxWidth:  38,
//MaxHeight: 38,
//}),
//widget.WidgetOpts.ToolTip(widget.NewToolTip(
//widget.ToolTipOpts.Content(removeToolTipContainer),
////widget.WidgetToolTipOpts.Delay(1*time.Second),
//widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
//widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
////When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
////They will default to what you see below if you do not provide them
//widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
//widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
//widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
//widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
//)),
//),
//// specify the images to sue
//widget.ButtonOpts.Image(cutils.ButtonImageFromImage(cutils.Images.Get(cutils.CrossImageKey))),

//// add a handler that reacts to clicking the button
//widget.ButtonOpts.ClickedHandler(func() func(args *widget.ButtonClickedEventArgs) {
//return func(args *widget.ButtonClickedEventArgs) {
//cp := hs.game.Store.Lines.FindCurrentPlayer()
//otm := hs.GetState().(HUDState).OpenTowerMenu
//actionDispatcher.RemoveTower(cp.ID, otm.ID)
//actionDispatcher.CloseTowerMenu()
//}
//}()),
//)
//towerMenuC.AddChild(rbtn)

//// Update button
//updateToolTipContainer1 := widget.NewContainer(
//widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
//widget.ContainerOpts.AutoDisableChildren(),
//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
//)

//updateToolTxt1 := widget.NewText(
//widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
//widget.TextOpts.Text(fmt.Sprintf(towerUpdateToolTipTmpl, 0, 0.0, 0.0, 0.0, updateTowerKeybind1), cutils.SmallFont, cutils.White),
//widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
//)
//hs.towerUpdateToolTip1 = updateToolTxt1
//updateToolTipContainer1.AddChild(updateToolTxt1)

//ubtn1 := widget.NewButton(
//// set general widget options
//widget.ButtonOpts.WidgetOpts(
//widget.WidgetOpts.LayoutData(widget.GridLayoutData{
//MaxWidth:  38,
//MaxHeight: 38,
//}),
//widget.WidgetOpts.ToolTip(widget.NewToolTip(
//widget.ToolTipOpts.Content(updateToolTipContainer1),
////widget.WidgetToolTipOpts.Delay(1*time.Second),
//widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
//widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
////When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
////They will default to what you see below if you do not provide them
//widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
//widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
//widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
//widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
//)),
//),
//// specify the images to sue
//widget.ButtonOpts.Image(cutils.ButtonImageFromImage(cutils.Images.Get(cutils.ArrowImageKey))),

//// add a handler that reacts to clicking the button
//widget.ButtonOpts.ClickedHandler(func() func(args *widget.ButtonClickedEventArgs) {
//return func(args *widget.ButtonClickedEventArgs) {
//cp := hs.game.Store.Lines.FindCurrentPlayer()
//// I know which is the current open one
//// I'll have 2 buttons so I can know the position
//// that it was selected to update
//tomid := hs.GetState().(HUDState).OpenTowerMenu.ID
//actionDispatcher.UpdateTower(cp.ID, tomid, "TODO")
//}
//}()),
//)

//updateToolTipContainer2 := widget.NewContainer(
//widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
//widget.ContainerOpts.AutoDisableChildren(),
//widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
//)

//updateToolTxt2 := widget.NewText(
//widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
//widget.TextOpts.Text(fmt.Sprintf(towerUpdateToolTipTmpl, 0, 0.0, 0.0, 0.0, updateTowerKeybind2), cutils.SmallFont, cutils.White),
//widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
//)
//hs.towerUpdateToolTip2 = updateToolTxt2
//updateToolTipContainer2.AddChild(updateToolTxt2)
//ubtn2 := widget.NewButton(
//// set general widget options
//widget.ButtonOpts.WidgetOpts(
//widget.WidgetOpts.LayoutData(widget.GridLayoutData{
//MaxWidth:  38,
//MaxHeight: 38,
//}),
//widget.WidgetOpts.ToolTip(widget.NewToolTip(
//widget.ToolTipOpts.Content(updateToolTipContainer2),
////widget.WidgetToolTipOpts.Delay(1*time.Second),
//widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
//widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
////When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
////They will default to what you see below if you do not provide them
//widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
//widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
//widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
//widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
//)),
//),
//// specify the images to sue
//widget.ButtonOpts.Image(cutils.ButtonImageFromImage(cutils.Images.Get(cutils.ArrowImageKey))),

//// add a handler that reacts to clicking the button
//widget.ButtonOpts.ClickedHandler(func() func(args *widget.ButtonClickedEventArgs) {
//return func(args *widget.ButtonClickedEventArgs) {
//cp := hs.game.Store.Lines.FindCurrentPlayer()
//// I know which is the current open one //// I'll have 2 buttons so I can know the position //// that it was selected to update //tomid := hs.GetState().(HUDState).OpenTowerMenu.ID //actionDispatcher.UpdateTower(cp.ID, tomid, "TODO")
//}
//}()),
//)
//towerMenuC.AddChild(ubtn1)
//towerMenuC.AddChild(ubtn2)
//bottomLeftContainer.AddChild(towerMenuC)
//hs.bottomLeftContainer = bottomLeftContainer
//hs.towerMenuContainer = towerMenuC
//hs.towerUpdateButton1 = ubtn1
//hs.towerUpdateButton2 = ubtn2

//return bottomLeftContainer
//}

func (hs *HUDStore) loadModal(w *widget.Window) {
	if hs.ui.IsWindowOpen(w) {
		return
	}
	//Get the preferred size of the content
	x, y := w.Contents.PreferredSize()
	//Create a rect with the preferred size of the content
	r := stdimage.Rect(0, 0, x, y)

	uirect := hs.ui.Container.GetWidget().Rect
	uix, uiy := uirect.Dx(), uirect.Dy()
	//Use the Add method to move the window to the specified point
	r = r.Add(stdimage.Point{uix/2 - x/2, uiy/2 - y/2})
	//Set the windows location to the rect.
	w.SetLocation(r)
	//Add the window to the UI.
	//Note: If the window is already added, this will just move the window and not add a duplicate.
	hs.ui.AddWindow(w)
}

func (hs *HUDStore) closeModal(w *widget.Window) {
	if !hs.ui.IsWindowOpen(w) {
		return
	}
	w.Close()
}

// towerToolTip return the container and the labels in this order:
// * title
// * damage
// * range
// * health
// * description
func (hs *HUDStore) towerToolTip(t *tower.Tower, kb ebiten.Key) (*widget.ToolTip, *widget.Text, *widget.Text, *widget.Text, *widget.Text, *widget.Text, *widget.Text) {
	tooltipC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(20)),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.ToolTipBGKey, 8, 8, !isPressed)),
	)

	tooltipDetailsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			//widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Columns(1),
			//Define how much padding to inset the child content
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(10, 0),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.AutoDisableChildren(),
	)

	tooltipDetailsRows := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			//widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Columns(2),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 0),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false, false}, []bool{false, false}),
		)),
	)

	ttTitleTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.Text(fmt.Sprintf(unitToolTipTitleTmpl, t.Name(), kb), cutils.SmallFont, cutils.White),
	)

	goldIconC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	ttGoldGW := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.GoldIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)
	goldIconC.AddChild(ttGoldGW)

	ttGoldTxtW := widget.NewText(
		widget.TextOpts.Text(fmt.Sprint(t.Gold), cutils.SFont20, cutils.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	damageIconC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	ttDamageGW := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.DamageIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)
	damageIconC.AddChild(ttDamageGW)

	ttDamageTxtW := widget.NewText(
		widget.TextOpts.Text(fmt.Sprint(t.Damage), cutils.SFont20, cutils.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	rangeIconC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	ttRangeGW := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.RangeIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)
	rangeIconC.AddChild(ttRangeGW)

	ttRangeTxtW := widget.NewText(
		widget.TextOpts.Text(fmt.Sprint(t.Range), cutils.SFont20, cutils.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	healthIconC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	ttHealthGW := widget.NewGraphic(
		widget.GraphicOpts.Image(cutils.Images.Get(cutils.HeartIconKey)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)
	healthIconC.AddChild(ttHealthGW)

	ttHealthTxtW := widget.NewText(
		widget.TextOpts.Text(fmt.Sprint(t.Health), cutils.SFont20, cutils.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)
	ttDescriptionTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.Text("Description:", cutils.SFont20, cutils.White),
	)
	ttDescriptionContentTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.Text(t.Description(), cutils.SFont20, cutils.White),
	)

	tooltipDetailsRows.AddChild(
		goldIconC,
		ttGoldTxtW,

		damageIconC,
		ttDamageTxtW,

		rangeIconC,
		ttRangeTxtW,

		healthIconC,
		ttHealthTxtW,
	)

	tooltipDetailsC.AddChild(
		tooltipDetailsRows,
	)

	tooltipC.AddChild(
		ttTitleTxt,
		tooltipDetailsC,
		ttDescriptionTxt,
		ttDescriptionContentTxt,
	)

	return widget.NewToolTip(
		widget.ToolTipOpts.Content(tooltipC),
		//widget.WidgetToolTipOpts.Delay(1*time.Second),
		widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
		widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
		//When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
		//They will default to what you see below if you do not provide them
		widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
		widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
		widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
		widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
	), ttTitleTxt, ttGoldTxtW, ttDamageTxtW, ttRangeTxtW, ttHealthTxtW, ttDescriptionContentTxt
}

// simpleTooltip returns a simple tooltip with a title and description, meant for the simple hover information
// it returns the tooltip and the title and descriptions widets
func (hs *HUDStore) simpleTooltip(title, description string) (*widget.ToolTip, *widget.Text, *widget.Text) {
	tooltipC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(20)),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.ToolTipBGKey, 8, 8, !isPressed)),
	)

	ttTitleTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.Text(title, cutils.SmallFont, cutils.White),
	)

	ttDescriptionContentTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.Text(description, cutils.SFont20, cutils.White),
	)

	tooltipC.AddChild(
		ttTitleTxt,
		ttDescriptionContentTxt,
	)

	return widget.NewToolTip(
		widget.ToolTipOpts.Content(tooltipC),
		//widget.WidgetToolTipOpts.Delay(1*time.Second),
		widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
		widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
		//When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
		//They will default to what you see below if you do not provide them
		widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
		widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
		widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
		widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
	), ttTitleTxt, ttDescriptionContentTxt
}
