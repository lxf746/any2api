var F = Object.defineProperty;
var G = (t, e, n) =>
  e in t
    ? F(t, e, { enumerable: !0, configurable: !0, writable: !0, value: n })
    : (t[e] = n);
var f = (t, e, n) => G(t, typeof e != "symbol" ? e + "" : e, n);
function H() {}
const _t = (t) => t;
function I(t, e) {
  for (const n in e) t[n] = e[n];
  return t;
}
function ht(t) {
  return (
    !!t &&
    (typeof t == "object" || typeof t == "function") &&
    typeof t.then == "function"
  );
}
function z(t) {
  return t();
}
function dt() {
  return Object.create(null);
}
function U(t) {
  t.forEach(z);
}
function W(t) {
  return typeof t == "function";
}
function mt(t, e) {
  return t != t
    ? e == e
    : t !== e || (t && typeof t == "object") || typeof t == "function";
}
let p;
function pt(t, e) {
  return t === e
    ? !0
    : (p || (p = document.createElement("a")), (p.href = e), t === p.href);
}
function yt(t, e) {
  return t != t ? e == e : t !== e;
}
function gt(t) {
  return Object.keys(t).length === 0;
}
function L(t, ...e) {
  if (t == null) {
    for (const i of e) i(void 0);
    return H;
  }
  const n = t.subscribe(...e);
  return n.unsubscribe ? () => n.unsubscribe() : n;
}
function bt(t) {
  let e;
  return (L(t, (n) => (e = n))(), e);
}
function xt(t, e, n) {
  t.$$.on_destroy.push(L(e, n));
}
function Et(t, e, n, i) {
  if (t) {
    const s = S(t, e, n, i);
    return t[0](s);
  }
}
function S(t, e, n, i) {
  return t[1] && i ? I(n.ctx.slice(), t[1](i(e))) : n.ctx;
}
function wt(t, e, n, i) {
  if (t[2] && i) {
    const s = t[2](i(n));
    if (e.dirty === void 0) return s;
    if (typeof s == "object") {
      const o = [],
        r = Math.max(e.dirty.length, s.length);
      for (let l = 0; l < r; l += 1) o[l] = e.dirty[l] | s[l];
      return o;
    }
    return e.dirty | s;
  }
  return e.dirty;
}
function Tt(t, e, n, i, s, o) {
  if (s) {
    const r = S(e, n, i, o);
    t.p(r, s);
  }
}
function Nt(t) {
  if (t.ctx.length > 32) {
    const e = [],
      n = t.ctx.length / 32;
    for (let i = 0; i < n; i++) e[i] = -1;
    return e;
  }
  return -1;
}
function kt(t) {
  const e = {};
  for (const n in t) n[0] !== "$" && (e[n] = t[n]);
  return e;
}
function vt(t, e) {
  const n = {};
  e = new Set(e);
  for (const i in t) !e.has(i) && i[0] !== "$" && (n[i] = t[i]);
  return n;
}
function At(t) {
  const e = {};
  for (const n in t) e[n] = !0;
  return e;
}
function Ct(t) {
  return t ?? "";
}
function jt(t, e, n) {
  return (t.set(n), e);
}
function Dt(t) {
  return t && W(t.destroy) ? t.destroy : H;
}
function Ht(t) {
  const e = typeof t == "string" && t.match(/^\s*(-?[\d.]+)([^\s]*)\s*$/);
  return e ? [parseFloat(e[1]), e[2] || "px"] : [t, "px"];
}
let g = !1;
function Lt() {
  g = !0;
}
function St() {
  g = !1;
}
function J(t, e, n, i) {
  for (; t < e; ) {
    const s = t + ((e - t) >> 1);
    n(s) <= i ? (t = s + 1) : (e = s);
  }
  return t;
}
function K(t) {
  if (t.hydrate_init) return;
  t.hydrate_init = !0;
  let e = t.childNodes;
  if (t.nodeName === "HEAD") {
    const c = [];
    for (let a = 0; a < e.length; a++) {
      const u = e[a];
      u.claim_order !== void 0 && c.push(u);
    }
    e = c;
  }
  const n = new Int32Array(e.length + 1),
    i = new Int32Array(e.length);
  n[0] = -1;
  let s = 0;
  for (let c = 0; c < e.length; c++) {
    const a = e[c].claim_order,
      u =
        (s > 0 && e[n[s]].claim_order <= a
          ? s + 1
          : J(1, s, (R) => e[n[R]].claim_order, a)) - 1;
    i[c] = n[u] + 1;
    const C = u + 1;
    ((n[C] = c), (s = Math.max(C, s)));
  }
  const o = [],
    r = [];
  let l = e.length - 1;
  for (let c = n[s] + 1; c != 0; c = i[c - 1]) {
    for (o.push(e[c - 1]); l >= c; l--) r.push(e[l]);
    l--;
  }
  for (; l >= 0; l--) r.push(e[l]);
  (o.reverse(), r.sort((c, a) => c.claim_order - a.claim_order));
  for (let c = 0, a = 0; c < r.length; c++) {
    for (; a < o.length && r[c].claim_order >= o[a].claim_order; ) a++;
    const u = a < o.length ? o[a] : null;
    t.insertBefore(r[c], u);
  }
}
function Q(t, e) {
  t.appendChild(e);
}
function V(t) {
  if (!t) return document;
  const e = t.getRootNode ? t.getRootNode() : t.ownerDocument;
  return e && e.host ? e : t.ownerDocument;
}
function Mt(t) {
  const e = k("style");
  return ((e.textContent = "/* empty */"), X(V(t), e), e.sheet);
}
function X(t, e) {
  return (Q(t.head || t, e), e.sheet);
}
function Y(t, e) {
  if (g) {
    for (
      K(t),
        (t.actual_end_child === void 0 ||
          (t.actual_end_child !== null &&
            t.actual_end_child.parentNode !== t)) &&
          (t.actual_end_child = t.firstChild);
      t.actual_end_child !== null && t.actual_end_child.claim_order === void 0;
    )
      t.actual_end_child = t.actual_end_child.nextSibling;
    e !== t.actual_end_child
      ? (e.claim_order !== void 0 || e.parentNode !== t) &&
        t.insertBefore(e, t.actual_end_child)
      : (t.actual_end_child = e.nextSibling);
  } else (e.parentNode !== t || e.nextSibling !== null) && t.appendChild(e);
}
function Z(t, e, n) {
  t.insertBefore(e, n || null);
}
function $(t, e, n) {
  g && !n
    ? Y(t, e)
    : (e.parentNode !== t || e.nextSibling != n) &&
      t.insertBefore(e, n || null);
}
function w(t) {
  t.parentNode && t.parentNode.removeChild(t);
}
function Pt(t, e) {
  for (let n = 0; n < t.length; n += 1) t[n] && t[n].d(e);
}
function k(t) {
  return document.createElement(t);
}
function M(t) {
  return document.createElementNS("http://www.w3.org/2000/svg", t);
}
function v(t) {
  return document.createTextNode(t);
}
function Ot() {
  return v(" ");
}
function qt() {
  return v("");
}
function Bt(t, e, n, i) {
  return (t.addEventListener(e, n, i), () => t.removeEventListener(e, n, i));
}
function Rt(t) {
  return function (e) {
    return (e.preventDefault(), t.call(this, e));
  };
}
function Ft(t) {
  return function (e) {
    return (e.stopPropagation(), t.call(this, e));
  };
}
function Gt(t) {
  return function (e) {
    e.target === this && t.call(this, e);
  };
}
function A(t, e, n) {
  n == null
    ? t.removeAttribute(e)
    : t.getAttribute(e) !== n && t.setAttribute(e, n);
}
const tt = ["width", "height"];
function et(t, e) {
  const n = Object.getOwnPropertyDescriptors(t.__proto__);
  for (const i in e)
    e[i] == null
      ? t.removeAttribute(i)
      : i === "style"
        ? (t.style.cssText = e[i])
        : i === "__value"
          ? (t.value = t[i] = e[i])
          : n[i] && n[i].set && tt.indexOf(i) === -1
            ? (t[i] = e[i])
            : A(t, i, e[i]);
}
function It(t, e) {
  for (const n in e) A(t, n, e[n]);
}
function nt(t, e) {
  Object.keys(e).forEach((n) => {
    it(t, n, e[n]);
  });
}
function it(t, e, n) {
  const i = e.toLowerCase();
  i in t
    ? (t[i] = typeof t[i] == "boolean" && n === "" ? !0 : n)
    : e in t
      ? (t[e] = typeof t[e] == "boolean" && n === "" ? !0 : n)
      : A(t, e, n);
}
function zt(t) {
  return /-/.test(t) ? nt : et;
}
function Ut(t) {
  return t.dataset.svelteH;
}
function Wt(t) {
  let e;
  return {
    p(...n) {
      ((e = n), e.forEach((i) => t.push(i)));
    },
    r() {
      e.forEach((n) => t.splice(t.indexOf(n), 1));
    },
  };
}
function Jt(t) {
  return t === "" ? null : +t;
}
function Kt(t) {
  return Array.from(t.childNodes);
}
function P(t) {
  t.claim_info === void 0 &&
    (t.claim_info = { last_index: 0, total_claimed: 0 });
}
function O(t, e, n, i, s = !1) {
  P(t);
  const o = (() => {
    for (let r = t.claim_info.last_index; r < t.length; r++) {
      const l = t[r];
      if (e(l)) {
        const c = n(l);
        return (
          c === void 0 ? t.splice(r, 1) : (t[r] = c),
          s || (t.claim_info.last_index = r),
          l
        );
      }
    }
    for (let r = t.claim_info.last_index - 1; r >= 0; r--) {
      const l = t[r];
      if (e(l)) {
        const c = n(l);
        return (
          c === void 0 ? t.splice(r, 1) : (t[r] = c),
          s
            ? c === void 0 && t.claim_info.last_index--
            : (t.claim_info.last_index = r),
          l
        );
      }
    }
    return i();
  })();
  return (
    (o.claim_order = t.claim_info.total_claimed),
    (t.claim_info.total_claimed += 1),
    o
  );
}
function q(t, e, n, i) {
  return O(
    t,
    (s) => s.nodeName === e,
    (s) => {
      const o = [];
      for (let r = 0; r < s.attributes.length; r++) {
        const l = s.attributes[r];
        n[l.name] || o.push(l.name);
      }
      o.forEach((r) => s.removeAttribute(r));
    },
    () => i(e),
  );
}
function Qt(t, e, n) {
  return q(t, e, n, k);
}
function Vt(t, e, n) {
  return q(t, e, n, M);
}
function st(t, e) {
  return O(
    t,
    (n) => n.nodeType === 3,
    (n) => {
      const i = "" + e;
      if (n.data.startsWith(i)) {
        if (n.data.length !== i.length) return n.splitText(i.length);
      } else n.data = i;
    },
    () => v(e),
    !0,
  );
}
function Xt(t) {
  return st(t, " ");
}
function j(t, e, n) {
  for (let i = n; i < t.length; i += 1) {
    const s = t[i];
    if (s.nodeType === 8 && s.textContent.trim() === e) return i;
  }
  return -1;
}
function Yt(t, e) {
  const n = j(t, "HTML_TAG_START", 0),
    i = j(t, "HTML_TAG_END", n + 1);
  if (n === -1 || i === -1) return new b(e);
  P(t);
  const s = t.splice(n, i - n + 1);
  (w(s[0]), w(s[s.length - 1]));
  const o = s.slice(1, s.length - 1);
  if (o.length === 0) return new b(e);
  for (const r of o)
    ((r.claim_order = t.claim_info.total_claimed),
      (t.claim_info.total_claimed += 1));
  return new b(e, o);
}
function Zt(t, e) {
  ((e = "" + e), t.data !== e && (t.data = e));
}
function $t(t, e) {
  ((e = "" + e), t.wholeText !== e && (t.data = e));
}
function te(t, e) {
  t.value = e ?? "";
}
function ee(t, e, n, i) {
  n == null ? t.style.removeProperty(e) : t.style.setProperty(e, n, "");
}
function ne(t, e, n) {
  for (let i = 0; i < t.options.length; i += 1) {
    const s = t.options[i];
    if (s.__value === e) {
      s.selected = !0;
      return;
    }
  }
  (!n || e !== void 0) && (t.selectedIndex = -1);
}
function ie(t) {
  const e = t.querySelector(":checked");
  return e && e.__value;
}
function se(t, e, n) {
  t.classList.toggle(e, !!n);
}
function rt(t, e, { bubbles: n = !1, cancelable: i = !1 } = {}) {
  return new CustomEvent(t, { detail: e, bubbles: n, cancelable: i });
}
function re(t, e) {
  const n = [];
  let i = 0;
  for (const s of e.childNodes)
    if (s.nodeType === 8) {
      const o = s.textContent.trim();
      o === `HEAD_${t}_END`
        ? ((i -= 1), n.push(s))
        : o === `HEAD_${t}_START` && ((i += 1), n.push(s));
    } else i > 0 && n.push(s);
  return n;
}
class ct {
  constructor(e = !1) {
    f(this, "is_svg", !1);
    f(this, "e");
    f(this, "n");
    f(this, "t");
    f(this, "a");
    ((this.is_svg = e), (this.e = this.n = null));
  }
  c(e) {
    this.h(e);
  }
  m(e, n, i = null) {
    (this.e ||
      (this.is_svg
        ? (this.e = M(n.nodeName))
        : (this.e = k(n.nodeType === 11 ? "TEMPLATE" : n.nodeName)),
      (this.t = n.tagName !== "TEMPLATE" ? n : n.content),
      this.c(e)),
      this.i(i));
  }
  h(e) {
    ((this.e.innerHTML = e),
      (this.n = Array.from(
        this.e.nodeName === "TEMPLATE"
          ? this.e.content.childNodes
          : this.e.childNodes,
      )));
  }
  i(e) {
    for (let n = 0; n < this.n.length; n += 1) Z(this.t, this.n[n], e);
  }
  p(e) {
    (this.d(), this.h(e), this.i(this.a));
  }
  d() {
    this.n.forEach(w);
  }
}
class b extends ct {
  constructor(n = !1, i) {
    super(n);
    f(this, "l");
    ((this.e = this.n = null), (this.l = i));
  }
  c(n) {
    this.l ? (this.n = this.l) : super.c(n);
  }
  i(n) {
    for (let i = 0; i < this.n.length; i += 1) $(this.t, this.n[i], n);
  }
}
function ce(t, e) {
  return new t(e);
}
let y;
function x(t) {
  y = t;
}
function _() {
  if (!y) throw new Error("Function called outside component initialization");
  return y;
}
function oe(t) {
  _().$$.on_mount.push(t);
}
function le(t) {
  _().$$.after_update.push(t);
}
function ae(t) {
  _().$$.on_destroy.push(t);
}
function ue() {
  const t = _();
  return (e, n, { cancelable: i = !1 } = {}) => {
    const s = t.$$.callbacks[e];
    if (s) {
      const o = rt(e, n, { cancelable: i });
      return (
        s.slice().forEach((r) => {
          r.call(t, o);
        }),
        !o.defaultPrevented
      );
    }
    return !0;
  };
}
function fe(t, e) {
  return (_().$$.context.set(t, e), e);
}
function _e(t) {
  return _().$$.context.get(t);
}
function he(t) {
  return _().$$.context.has(t);
}
function de(t, e) {
  const n = t.$$.callbacks[e.type];
  n && n.slice().forEach((i) => i.call(this, e));
}
const m = [],
  D = [];
let d = [];
const T = [],
  B = Promise.resolve();
let N = !1;
function ot() {
  N || ((N = !0), B.then(at));
}
function me() {
  return (ot(), B);
}
function lt(t) {
  d.push(t);
}
function pe(t) {
  T.push(t);
}
const E = new Set();
let h = 0;
function at() {
  if (h !== 0) return;
  const t = y;
  do {
    try {
      for (; h < m.length; ) {
        const e = m[h];
        (h++, x(e), ut(e.$$));
      }
    } catch (e) {
      throw ((m.length = 0), (h = 0), e);
    }
    for (x(null), m.length = 0, h = 0; D.length; ) D.pop()();
    for (let e = 0; e < d.length; e += 1) {
      const n = d[e];
      E.has(n) || (E.add(n), n());
    }
    d.length = 0;
  } while (m.length);
  for (; T.length; ) T.pop()();
  ((N = !1), E.clear(), x(t));
}
function ut(t) {
  if (t.fragment !== null) {
    (t.update(), U(t.before_update));
    const e = t.dirty;
    ((t.dirty = [-1]),
      t.fragment && t.fragment.p(t.ctx, e),
      t.after_update.forEach(lt));
  }
}
function ye(t) {
  const e = [],
    n = [];
  (d.forEach((i) => (t.indexOf(i) === -1 ? e.push(i) : n.push(i))),
    n.forEach((i) => i()),
    (d = e));
}
export {
  _t as $,
  Rt as A,
  pe as B,
  te as C,
  me as D,
  U as E,
  M as F,
  Vt as G,
  Pt as H,
  Gt as I,
  ue as J,
  I as K,
  et as L,
  vt as M,
  kt as N,
  Ut as O,
  de as P,
  ee as Q,
  pt as R,
  L as S,
  W as T,
  bt as U,
  ce as V,
  ie as W,
  lt as X,
  ne as Y,
  It as Z,
  Ct as _,
  Kt as a,
  Ht as a0,
  Dt as a1,
  le as a2,
  fe as a3,
  jt as a4,
  Jt as a5,
  V as a6,
  Mt as a7,
  rt as a8,
  x as a9,
  dt as aa,
  at as ab,
  gt as ac,
  ye as ad,
  y as ae,
  z as af,
  m as ag,
  ot as ah,
  Lt as ai,
  St as aj,
  se as ak,
  Wt as al,
  Ft as am,
  yt as an,
  b as ao,
  Yt as ap,
  zt as aq,
  ht as ar,
  _ as as,
  he as at,
  At as au,
  $t as av,
  st as b,
  Qt as c,
  w as d,
  k as e,
  A as f,
  Y as g,
  Zt as h,
  $ as i,
  xt as j,
  Ot as k,
  qt as l,
  re as m,
  H as n,
  Xt as o,
  _e as p,
  oe as q,
  ae as r,
  mt as s,
  v as t,
  Et as u,
  Bt as v,
  Tt as w,
  Nt as x,
  wt as y,
  D as z,
};
