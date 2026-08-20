#!/usr/bin/env python3
"""Render docs/demo/vscode-demo.gif — an editor-style animation of the goninja
loop: annotate a struct, and the generated resource appears beside it.

    pip install pillow pygments
    python3 docs/demo/render_vscode_demo.py

The palette is Lilac Nights (github.com/caspel26/lilac-nights), inlined here so
this script stays self-contained and reproducible on any machine. The generated
pane is a verbatim excerpt of examples/prototype/internal/api/book_generated.go
— it is read off disk, not retyped, so it cannot drift from real output.
"""

from pathlib import Path

from PIL import Image, ImageDraw, ImageFont
from pygments import lex
from pygments.lexers import GoLexer
from pygments.token import Token

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "docs" / "demo" / "vscode-demo.gif"
GENERATED = ROOT / "examples" / "prototype" / "internal" / "api" / "book_generated.go"

# --- Lilac Nights ------------------------------------------------------------
C = {
    "bg": "#211a2e",
    "chrome": "#15111f",
    "line_hl": "#2b2340",
    "gutter": "#716490",
    "gutter_active": "#d7b8ff",
    "fg": "#e6e0f5",
    "border": "#372c52",
    "keyword": "#b98cff",
    "punct": "#8a7dad",
    "attr": "#d7b8ff",
    "func": "#8ec5ff",
    "cls": "#ffcc80",
    "builtin_type": "#c3a279",
    "string": "#8fdc98",
    "number": "#ff9d7a",
    "param": "#f17ec1",
    "operator": "#5fe3d4",
    "comment": "#8a7dad",
    "tab_active_bg": "#211a2e",
    "tab_inactive_bg": "#15111f",
    "tab_active_fg": "#f2ecff",
    "tab_inactive_fg": "#716490",
    "tab_top": "#d7b8ff",
    "status_bg": "#15111f",
    "status_fg": "#c3bade",
}
TRAFFIC = ["#ff5f57", "#febc2e", "#28c840"]

# goninja brand backdrop: deep purple into near-black.
BG_TOP, BG_BOTTOM = (0x2A, 0x12, 0x45), (0x0D, 0x0C, 0x14)

S = 2  # supersample factor; the GIF is downscaled from this for clean edges
W, H = 1040 * S, 570 * S
PAD = 34 * S
BAR = 34 * S          # title bar
TABS = 36 * S         # tab strip
STATUS = 26 * S       # status bar
GUTTER = 54 * S
FS = 15 * S
LH = 23 * S
RADIUS = 11 * S
TERM_H = 132 * S      # integrated terminal panel

MENLO = "/System/Library/Fonts/Menlo.ttc"
SANS = "/System/Library/Fonts/Supplemental/Arial.ttf"

STYLES = {
    Token.Comment: ("comment", "italic"),
    Token.Keyword: ("keyword", "bold"),
    Token.Keyword.Type: ("builtin_type", "italic"),
    Token.Keyword.Declaration: ("keyword", "bold"),
    Token.Name: ("fg", "regular"),
    Token.Name.Attribute: ("attr", "regular"),
    Token.Name.Builtin: ("cls", "italic"),
    Token.Name.Class: ("cls", "bold"),
    Token.Name.Function: ("func", "regular"),
    Token.Name.Namespace: ("cls", "italic"),
    Token.Name.Parameter: ("param", "regular"),
    Token.Literal.Number: ("number", "regular"),
    Token.Literal.String: ("string", "regular"),
    Token.Operator: ("operator", "regular"),
    Token.Punctuation: ("punct", "regular"),
    Token.Text: ("fg", "regular"),
}
TYPE_NAMES = {"string", "bool", "int", "int8", "int16", "int32", "int64", "uint",
              "uint8", "uint16", "uint32", "uint64", "float32", "float64",
              "byte", "rune", "error", "any"}


def fonts():
    return {
        "regular": ImageFont.truetype(MENLO, FS, index=0),
        "bold": ImageFont.truetype(MENLO, FS, index=1),
        "italic": ImageFont.truetype(MENLO, FS, index=2),
        "ui": ImageFont.truetype(SANS, int(12.5 * S)),
        "ui_bold": ImageFont.truetype(SANS, int(12.5 * S)),
    }


def style_for(ttype):
    t = ttype
    while t is not None:
        if t in STYLES:
            key, st = STYLES[t]
            return C[key], st
        t = t.parent
    return C["fg"], "regular"


def tokenize(code):
    """Pygments tokens as (text, color, style) per line."""
    code = code.expandtabs(4)
    raw = [(t, v) for t, v in lex(code, GoLexer()) if v]
    lines, cur = [], []
    prev, bol = "", True
    for i, (ttype, value) in enumerate(raw):
        nxt = next((v2.strip() for _, v2 in raw[i + 1:] if v2.strip()), "")
        if ttype in Token.Name:
            plain = ttype in (Token.Name, Token.Name.Other)
            if prev == ".":
                ttype = Token.Name.Function if nxt == "(" else Token.Name.Attribute
            elif nxt == "(" and plain:
                ttype = Token.Name.Function
            elif value in TYPE_NAMES:
                ttype = Token.Keyword.Type
            elif plain and bol and (nxt in TYPE_NAMES or nxt == "[" or nxt[:1].isupper()):
                ttype = Token.Name  # a field declaration, painted plain
            elif value[:1].isupper() and not value.isupper() and plain:
                ttype = Token.Name.Class
        if value.strip():
            prev, bol = value.strip(), False
        color, st = style_for(ttype)
        for j, part in enumerate(value.split("\n")):
            if j:
                lines.append(cur)
                cur = []
                bol = True
            if part:
                cur.append((part, color, st))
    lines.append(cur)
    return lines


def backdrop():
    img = Image.new("RGB", (W, H))
    d = ImageDraw.Draw(img)
    for y in range(H):
        t = y / H
        d.line([(0, y), (W, y)], fill=tuple(
            int(a + (b - a) * t) for a, b in zip(BG_TOP, BG_BOTTOM)))
    return img


def rounded_mask(size, radius):
    m = Image.new("L", (size[0] * 2, size[1] * 2), 0)
    ImageDraw.Draw(m).rounded_rectangle(
        [0, 0, size[0] * 2 - 1, size[1] * 2 - 1], radius=radius * 2, fill=255)
    return m.resize(size, Image.LANCZOS)


def draw_window(f, tabs, active, lines, title, status_right, cursor_at=None,
                terminal=None, term_cursor=False):
    """One frame: the editor window with `lines` already tokenized.

    `terminal` is a list of (text, colour) rows for VS Code's integrated
    terminal panel; None hides the panel entirely.
    """
    win_w, win_h = W - PAD * 2, H - PAD * 2
    win = Image.new("RGB", (win_w, win_h), C["bg"])
    d = ImageDraw.Draw(win)

    # title bar
    d.rectangle([0, 0, win_w, BAR], fill=C["chrome"])
    for i, col in enumerate(TRAFFIC):
        cx = 18 * S + i * 19 * S
        r = 6 * S
        d.ellipse([cx - r, BAR // 2 - r, cx + r, BAR // 2 + r], fill=col)
    tw = d.textlength(title, font=f["ui"])
    d.text(((win_w - tw) / 2, BAR / 2 - 8 * S), title, font=f["ui"], fill=C["status_fg"])

    # tab strip
    d.rectangle([0, BAR, win_w, BAR + TABS], fill=C["tab_inactive_bg"])
    x = 0
    for name in tabs:
        twid = d.textlength(name, font=f["ui"]) + 40 * S
        is_active = name == active
        if is_active:
            d.rectangle([x, BAR, x + twid, BAR + TABS], fill=C["tab_active_bg"])
            d.rectangle([x, BAR, x + twid, BAR + 2 * S], fill=C["tab_top"])
        d.text((x + 20 * S, BAR + TABS / 2 - 8 * S), name, font=f["ui"],
               fill=C["tab_active_fg"] if is_active else C["tab_inactive_fg"])
        x += twid
    d.line([(0, BAR + TABS), (win_w, BAR + TABS)], fill=C["border"], width=1 * S)

    # code area (shortened when the terminal panel is open)
    term_h = TERM_H if terminal is not None else 0
    top = BAR + TABS + 10 * S
    body_h = win_h - top - STATUS - term_h
    max_rows = int(body_h // LH)
    view = lines[-max_rows:] if len(lines) > max_rows else lines

    for i, toks in enumerate(view):
        y = top + i * LH
        is_cursor_row = cursor_at is not None and i == len(view) - 1
        if is_cursor_row:
            d.rectangle([GUTTER, y - 3 * S, win_w - 78 * S, y + LH - 3 * S],
                        fill=C["line_hl"])
        num = str(i + 1 + max(0, len(lines) - max_rows))
        nw = d.textlength(num, font=f["regular"])
        d.text((GUTTER - 14 * S - nw, y), num, font=f["regular"],
               fill=C["gutter_active"] if is_cursor_row else C["gutter"])
        x = GUTTER + 8 * S
        for text, color, st in toks:
            d.text((x, y), text, font=f[st if st in f else "regular"], fill=color)
            x += d.textlength(text, font=f[st if st in f else "regular"])
        if is_cursor_row and cursor_at:
            d.rectangle([x, y + 1 * S, x + 2 * S, y + LH - 5 * S], fill=C["attr"])

    # minimap: abstract bars mirroring line lengths
    mm_x = win_w - 70 * S
    for i, toks in enumerate(view):
        ln = sum(len(t[0]) for t in toks)
        if not ln:
            continue
        wpx = min(58 * S, int(ln * 0.7 * S))
        yy = top + i * (LH * 0.34)
        if yy > win_h - STATUS - 8 * S:
            break
        d.rectangle([mm_x, yy, mm_x + wpx, yy + 2 * S], fill=C["border"])

    # integrated terminal panel
    if terminal is not None:
        ty = win_h - STATUS - term_h
        d.rectangle([0, ty, win_w, win_h - STATUS], fill=C["chrome"])
        d.line([(0, ty), (win_w, ty)], fill=C["border"], width=1 * S)
        # panel header: PROBLEMS  OUTPUT  TERMINAL
        hx = 16 * S
        for label in ("PROBLEMS", "OUTPUT", "TERMINAL"):
            active_panel = label == "TERMINAL"
            lw = d.textlength(label, font=f["ui"])
            d.text((hx, ty + 9 * S), label, font=f["ui"],
                   fill=C["tab_active_fg"] if active_panel else C["tab_inactive_fg"])
            if active_panel:
                d.rectangle([hx, ty + 27 * S, hx + lw, ty + 28 * S + 1 * S],
                            fill=C["tab_top"])
            hx += lw + 26 * S
        # rows
        ry = ty + 40 * S
        for text, color in terminal:
            d.text((16 * S, ry), text, font=f["regular"], fill=color)
            ry += LH
        if term_cursor and terminal:
            last, _ = terminal[-1]
            cw = d.textlength(last, font=f["regular"])
            d.rectangle([16 * S + cw + 2 * S, ry - LH + 2 * S,
                         16 * S + cw + 10 * S, ry - LH + FS + 2 * S],
                        fill=C["fg"])

    # status bar
    sy = win_h - STATUS
    d.rectangle([0, sy, win_w, win_h], fill=C["status_bg"])
    d.text((16 * S, sy + STATUS / 2 - 8 * S), "main", font=f["ui"],
           fill=C["status_fg"])
    d.text((92 * S, sy + STATUS / 2 - 8 * S), "Go 1.25", font=f["ui"],
           fill=C["status_fg"])
    rw = d.textlength(status_right, font=f["ui"])
    d.text((win_w - rw - 16 * S, sy + STATUS / 2 - 8 * S), status_right,
           font=f["ui"], fill=C["status_fg"])

    # composite onto the brand backdrop with rounded corners + shadow
    frame = backdrop()
    mask = rounded_mask((win_w, win_h), RADIUS)
    shadow = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    sd = ImageDraw.Draw(shadow)
    sd.rounded_rectangle([PAD, PAD + 6 * S, PAD + win_w, PAD + win_h + 6 * S],
                         radius=RADIUS, fill=(0, 0, 0, 90))
    frame = Image.alpha_composite(frame.convert("RGBA"), shadow).convert("RGB")
    frame.paste(win, (PAD, PAD), mask)
    return frame.resize((W // S, H // S), Image.LANCZOS)


# The command, and the output goninja actually prints (cmd/goninja/main.go:
# "goninja: generated %d model(s) into %s").
CMD = "goninja generate -models-import bookstore/models"
CMD_OUTPUT = "goninja: generated 2 model(s) into ./internal/api"
TERM_TITLE = "bash — bookstore"

MODEL_SRC = '''package models

// Book is the model. The goninja tags are the whole API definition.
type Book struct {
\tID        string    `gorm:"primaryKey" goninja:"list,retrieve"`
\tTitle     string    `goninja:"list,retrieve,create,update" validate:"required"`
\tAuthorID  string    `goninja:"list,retrieve,create,update,filter"`
\tPrice     float64   `goninja:"list,retrieve,create,update,filter"`
\tPublished bool      `goninja:"list,retrieve,create,update,filter"`
\tAuthor    Author    `goninja:"retrieve"`
}
'''


def generated_excerpt():
    """Verbatim slice of the real generated file: header, package, and BookList."""
    src = GENERATED.read_text().splitlines()
    head = src[:3]
    i = next(n for n, l in enumerate(src) if l.startswith("type BookList struct"))
    return "\n".join(head + ["", "// ...", ""] + src[i - 1:i + 8]) + "\n"


def main():
    f = fonts()
    frames, durations = [], []

    def add(img, ms):
        frames.append(img)
        durations.append(ms)

    tab1, tab2 = "book.go", "book_generated.go"
    title1 = "bookstore — book.go"
    title2 = "bookstore — book_generated.go"

    # 1. type the model, a few characters per frame
    step = 5
    for n in range(0, len(MODEL_SRC) + 1, step):
        partial = MODEL_SRC[:n]
        add(draw_window(f, [tab1], tab1, tokenize(partial), title1,
                        "Editing book.go", cursor_at=True), 40)
    add(draw_window(f, [tab1], tab1, tokenize(MODEL_SRC), title1,
                    "Editing book.go", cursor_at=True), 1400)

    model_lines = tokenize(MODEL_SRC)

    # 2. open the integrated terminal and type the generate command
    prompt = "$ "
    add(draw_window(f, [tab1], tab1, model_lines, title1, TERM_TITLE,
                    terminal=[(prompt, C["fg"])], term_cursor=True), 500)
    for n in range(1, len(CMD) + 1, 2):
        add(draw_window(f, [tab1], tab1, model_lines, title1, TERM_TITLE,
                        terminal=[(prompt + CMD[:n], C["fg"])],
                        term_cursor=True), 45)
    cmd_row = [(prompt + CMD, C["fg"])]
    add(draw_window(f, [tab1], tab1, model_lines, title1, TERM_TITLE,
                    terminal=cmd_row, term_cursor=True), 600)

    # 3. the real output, then the generated tab appears
    out_rows = cmd_row + [(CMD_OUTPUT, C["string"])]
    add(draw_window(f, [tab1], tab1, model_lines, title1, TERM_TITLE,
                    terminal=out_rows, term_cursor=False), 1300)

    gen = generated_excerpt()
    add(draw_window(f, [tab1, tab2], tab1, model_lines, title1,
                    TERM_TITLE, terminal=out_rows), 700)

    # 3. reveal the generated file, line by line
    glines = gen.split("\n")
    for k in range(1, len(glines) + 1):
        add(draw_window(f, [tab1, tab2], tab2, tokenize("\n".join(glines[:k])),
                        title2, "Generated — do not edit"), 110)
    add(draw_window(f, [tab1, tab2], tab2, tokenize(gen), title2,
                    "Generated — do not edit"), 2800)

    OUT.parent.mkdir(parents=True, exist_ok=True)
    # One shared adaptive palette: smaller file, and no colour flicker between
    # frames from per-frame quantisation.
    pal = frames[-1].quantize(colors=200, method=Image.MEDIANCUT)
    frames = [fr.quantize(palette=pal, dither=Image.NONE) for fr in frames]
    frames[0].save(OUT, save_all=True, append_images=frames[1:],
                   duration=durations, loop=0, optimize=True)
    kb = OUT.stat().st_size / 1024
    print(f"wrote {OUT.relative_to(ROOT)}  ({len(frames)} frames, {kb:.0f} KB)")


if __name__ == "__main__":
    main()
