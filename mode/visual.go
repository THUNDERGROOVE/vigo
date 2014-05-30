package mode

import (
	cmd "github.com/kisielk/vigo/commands"
	"github.com/kisielk/vigo/editor"
	"github.com/kisielk/vigo/view"
	"github.com/kisielk/vigo/utils"
	"github.com/nsf/termbox-go"
)

type visualMode struct {
	editor *editor.Editor
	count  string
	lineMode bool
}

func NewVisualMode(e *editor.Editor, lineMode bool) *visualMode {
	m := visualMode{editor: e, lineMode: lineMode}
	v := m.editor.ActiveView()
	c := v.Cursor()

	startPos, endPos := 0, 0

	if lineMode {
		m.editor.SetStatus("Visual Line")
	} else {
		startPos = c.Boffset
		endPos = c.Boffset

		m.editor.SetStatus("Visual")
	}

	viewTag := view.NewTag(
		c.LineNum,
		startPos,
		c.LineNum,
		endPos,
		termbox.ColorDefault,
		termbox.ColorDefault|termbox.AttrReverse,
	)

	v.SetVisualRange(&viewTag)

	return &m
}

func (m *visualMode) OnKey(ev *termbox.Event) {

	// Consequtive non-zero digits specify action multiplier;
	// accumulate and return. Accept zero only if it's
	// a non-starting character.
	if ('0' < ev.Ch && ev.Ch <= '9') || (ev.Ch == '0' && len(m.count) > 0) {
		m.count = m.count + string(ev.Ch)
		m.editor.SetStatus(m.count)
		return
	}
	count := utils.ParseCount(m.count)
	if count == 0 {
		count = 1
	}
	g := m.editor
	v := g.ActiveView()
	c := v.Cursor()
	vRange := v.VisualRange()

	startLine, startPos := vRange.StartPos()
	endLine, endPos := vRange.EndPos()

	switch ev.Key {
	case termbox.KeyEsc:
		m.editor.SetMode(NewNormalMode(m.editor))
	}

	switch ev.Ch {
	case 'h':
		if !m.lineMode {
			if c.Boffset >= startPos || c.LineNum != startLine {
				// cursor is right of the start of the selection
				// or anywhere on the line other that the start line
				// move the end of the selection left
				vRange.AdjustEndOffset(-count)
			} else {
				// cursor is anywhere on the start line
				// move the start of the selection left
				vRange.AdjustStartOffset(-count)
			}

			v.SetVisualRange(vRange)
		}
		g.Commands <- cmd.Repeat{cmd.MoveRune{Dir: cmd.Backward}, count}
	case 'j':
		if c.LineNum >= endLine {
			// cursor is below the end of the selection
			// move the end of the selection down
			vRange.AdjustEndLine(count)
		} else {
			// cursor is above the selection
			// move the start of the selection down
			vRange.AdjustStartLine(count)
		}

		// cursor is one line above the selection and further
		// along the line than the start of the selection
		// flip the start and end offsets of the selection
		if c.Boffset > endPos && c.LineNum == endLine -1 {
			vRange.FlipStartAndEndOffsets()
		}

		// cursor is on the same line as the entire selection, and left of the end of
		// the selection.
		// flip the offsets
		if c.LineNum == endLine && c.LineNum == startLine && c.Boffset < endPos {
			vRange.FlipStartAndEndOffsets()
		}

		v.SetVisualRange(vRange)
		g.Commands <- cmd.Repeat{cmd.MoveLine{Dir: cmd.Forward}, count}
	case 'k':
		if c.LineNum <= startLine {
			// cursor is above the start of the selection
			// move the start of the selection up
			vRange.AdjustStartLine(-count)
		} else {
			// cursor is below the start of the selection
			// move the end of the selection up
			vRange.AdjustEndLine(-count)
		}

		// cursor if right of the start of the selection and above
		// flip the start and end offsets of the selection
		if c.Boffset > startPos && c.LineNum <= startLine {
			vRange.FlipStartAndEndOffsets()
		}
		
		// cursor is left of the start of the selection
		// and one line below.
		// flip the offsets
		if c.Boffset < startPos && c.LineNum == startLine +1 {
			vRange.FlipStartAndEndOffsets()
		}

		v.SetVisualRange(vRange)
		g.Commands <- cmd.Repeat{cmd.MoveLine{Dir: cmd.Backward}, count}
	case 'l':
		if !m.lineMode {
			if c.Boffset >= endPos && c.LineNum == endLine {
				// cursor is right of the end of the selection and on the same line
				// move the end of the selection right
				vRange.AdjustEndOffset(count)
			} else {
				// cursor is left of the selection on any line
				// move the start of the selection right
				vRange.AdjustStartOffset(count)
			}

			v.SetVisualRange(vRange)
		}
		g.Commands <- cmd.Repeat{cmd.MoveRune{Dir: cmd.Forward}, count}
	case 'd':
		start, end := view.GetVisualSelection(v)
		v.Buffer().DeleteRange(start, end)
		m.editor.SetMode(NewNormalMode(m.editor))
	case 'v':
		m.editor.SetMode(NewNormalMode(m.editor))
	case 'V':
		if m.lineMode {
			m.editor.SetMode(NewNormalMode(m.editor))
		} else {
			v.VisualRange().SetStartOffset(0)
			v.VisualRange().SetEndOffset(len(c.Line.Data))
		}
	}

	// FIXME: there must be a better way of doing this
	// trigger a re-draw
	g.Resize()

	m.count = ""
}

func (m *visualMode) Exit() {
	v := m.editor.ActiveView()
	v.SetVisualRange(nil)
}