const ot = "Z.ai - Free AI Chatbot & Agent powered by GLM-5 & GLM-4.7",
  k = "",
  st = `${k}/api/v1`,
  it = `${k}/ollama`,
  at = `${k}/openai`,
  ct = `${k}/api/v1/audio`,
  ut = `${k}/api/v1/images`,
  lt = `${k}/api/v1/retrieval`,
  dt = "0.6.2",
  ft = "dev-build",
  ht = 1e3,
  _t = {
    ppt_composer: {
      text: "AI Slides",
      icon: "slides",
      color: "text-[#F07010]",
      bgColor: "hover:bg-[#F07010]/10",
    },
    ai_design: {
      text: "Magic Design",
      icon: "magic",
      color: "text-[#FF4098]",
      bgColor: "hover:bg-[#FF4098]/10",
    },
    web_dev: {
      text: "Full-Stack",
      icon: "code",
      color: " text-[#8945E9]",
      bgColor: "hover:bg-[#8945E9]/10",
    },
    grounding: {
      text: "Visual Positioning",
      icon: "grounding",
      color: " text-[#FF161A]",
      bgColor: "hover:bg-[#FF161A]/10",
    },
    ui_to_code: {
      text: "Webpage Replication",
      icon: "ui_to_code",
      color: " text-[#1296DB]",
      bgColor: "hover:bg-[#1296DB]/10",
    },
    deep_research: {
      text: "Deep Research",
      icon: "deep_research",
      color: " text-[#00cceb]",
      bgColor: "hover:bg-[#00cceb]/10",
    },
    visual_search: {
      text: "Visual Recognition",
      icon: "visual_search",
      color: " text-[#FF161A]",
      bgColor: "hover:bg-[#FF161A]/10",
    },
    smart_assistant: {
      text: "OCR Scan",
      icon: "information_scan",
      color: " text-[#FF8C00]",
      bgColor: "hover:bg-[#FF8C00]/10",
    },
    video_understanding: {
      text: "Video Understanding_2",
      icon: "video_understanding",
      color: " text-[#FF3300]",
      bgColor: "hover:bg-[#FF3300]/10",
    },
    edu_solver: {
      text: "Math Solve",
      icon: "edu_solver",
      color: " text-[#009944]",
      bgColor: "hover:bg-[#009944]/10",
    },
    doc_analysis: {
      text: "Visual Report",
      icon: "doc_analysis",
      color: " text-[#9c27b0]",
      bgColor: "hover:bg-[#9c27b0]/10",
    },
    smart_comparison: {
      text: "Smart Comparison",
      icon: "shopping",
      color: " text-[#1296db]",
      bgColor: "hover:bg-[#1296db]/10",
    },
    ui2code: {
      text: "UI Replication",
      icon: "ui_to_code",
      color: " text-[#1296DB]",
      bgColor: "hover:bg-[#1296DB]/10",
    },
    write: {
      text: "Writing",
      icon: "write",
      color: " text-[#009944]",
      bgColor: "hover:bg-[#009944]/10",
    },
    analysis: {
      text: "Data Insight",
      icon: "analysis",
      color: " text-[#9c27b0]",
      bgColor: "hover:bg-[#9c27b0]/10",
    },
  };
var et = ((s) => (
  (s.PPT_UPDATE = "ppt:update"),
  (s.PPT_SELECTED = "ppt:selected"),
  (s.PPT_SHOW_TYPE = "ppt:show_type"),
  (s.PPT_WORK_START = "ppt:work:start"),
  (s.PPT_WORK_PROCESSING = "ppt:work:processing"),
  (s.PPT_WORK_COMPLETE = "ppt:work:complete"),
  (s.WORKSPACE_PROCESS_ON = "workspace:process:on"),
  (s.WORKSPACE_PREVIEW_ON = "workspace:preview:on"),
  (s.WORKSPACE_PREVIEW_SWITCH = "workspace:preview:switch"),
  (s.WORKSPACE_PREVIEW_RELOAD = "workspace:preview:reload"),
  (s.WORKSPACE_CHECK_STATUS = "workspace:check:status"),
  (s.WORKSPACE_WRITE_DIFF = "workspace:write:diff"),
  (s.WORKSPACE_TERMINAL_UPDATE = "workspace:terminal:update"),
  (s.WORKSPACE_TERMINAL_HIDE = "workspace:terminal:hide"),
  (s.SUBMIT_PROMPT = "submit:prompt"),
  (s.THINKING_CHAIN_STAGE_UPDATE = "thinking_chain:stage:update"),
  (s.SHOW_MESSAGE_LOADING = "message_loading:show"),
  (s.COMPLETION_FINISH = "completion:finish"),
  (s.USER_INPUT_PROMPT = "user_input:prompt"),
  (s.MODE_CHANGE = "mode:change"),
  (s.LAST_TOOL_CALL_START = "last_tool_call:start"),
  (s.LAST_TOOL_CALL_UPDATE = "last_tool_call:update"),
  (s.LAST_TOOL_CALL_COMPLETE = "last_tool_call:complete"),
  (s.SHOW_WRITE_BY_SECTION = "show_write_by_section"),
  (s.SHOW_VISUAL_CHARTS_GEN_REPORT = "show_visual_charts_gen_report"),
  s
))(et || {});
const pt = {
    ADVANCED_SEARCH_OPEN: "advanced_search_open",
    GUIDED_ADVANCED_SEARCH_TIMES: "guided_advanced_search_times",
    MESSAGE_VERSION: "message_version",
    LAST_MODE: "last_mode",
    SELECTED_MODELS: "selectedModels",
  },
  gt = "/home/z/my-project/",
  mt = ["general_agent", "write", "analysis", "web_dev"],
  St = ["general_agent", "write", "analysis"],
  $t = [
    "md",
    "txt",
    "html",
    "pdf",
    "doc",
    "docx",
    "csv",
    "xls",
    "xlsx",
    "pptx",
    "ppt",
    "png",
    "jpg",
    "jpeg",
    "bmp",
    "gif",
    "mp3",
    "mp4",
    "wav",
  ],
  Ot = {
    SENSITIVE: "SENSITIVE",
    MODEL_CONCURRENCY_LIMIT: "MODEL_CONCURRENCY_LIMIT",
    INTERNAL_ERROR: "INTERNAL_ERROR",
  };
var Q =
  typeof globalThis < "u"
    ? globalThis
    : typeof window < "u"
      ? window
      : typeof global < "u"
        ? global
        : typeof self < "u"
          ? self
          : {};
function X(s) {
  return s && s.__esModule && Object.prototype.hasOwnProperty.call(s, "default")
    ? s.default
    : s;
}
function Mt(s) {
  if (s.__esModule) return s;
  var N = s.default;
  if (typeof N == "function") {
    var g = function m() {
      return this instanceof m
        ? Reflect.construct(N, arguments, this.constructor)
        : N.apply(this, arguments);
    };
    g.prototype = N.prototype;
  } else g = {};
  return (
    Object.defineProperty(g, "__esModule", { value: !0 }),
    Object.keys(s).forEach(function (m) {
      var b = Object.getOwnPropertyDescriptor(s, m);
      Object.defineProperty(
        g,
        m,
        b.get
          ? b
          : {
              enumerable: !0,
              get: function () {
                return s[m];
              },
            },
      );
    }),
    g
  );
}
var B = { exports: {} };
(function (s, N) {
  (function (g, m) {
    s.exports = m();
  })(Q, function () {
    var g = 1e3,
      m = 6e4,
      b = 36e5,
      T = "millisecond",
      y = "second",
      D = "minute",
      C = "hour",
      l = "day",
      M = "week",
      _ = "month",
      W = "quarter",
      E = "year",
      v = "date",
      w = "Invalid Date",
      H =
        /^(\d{4})[-/]?(\d{1,2})?[-/]?(\d{0,2})[Tt\s]*(\d{1,2})?:?(\d{1,2})?:?(\d{1,2})?[.:]?(\d+)?$/,
      Y =
        /\[([^\]]+)]|Y{1,4}|M{1,4}|D{1,2}|d{1,4}|H{1,2}|h{1,2}|a|A|m{1,2}|s{1,2}|Z{1,2}|SSS/g,
      j = {
        name: "en",
        weekdays:
          "Sunday_Monday_Tuesday_Wednesday_Thursday_Friday_Saturday".split("_"),
        months:
          "January_February_March_April_May_June_July_August_September_October_November_December".split(
            "_",
          ),
        ordinal: function (o) {
          var r = ["th", "st", "nd", "rd"],
            t = o % 100;
          return "[" + o + (r[(t - 20) % 10] || r[t] || r[0]) + "]";
        },
      },
      K = function (o, r, t) {
        var n = String(o);
        return !n || n.length >= r
          ? o
          : "" + Array(r + 1 - n.length).join(t) + o;
      },
      I = {
        s: K,
        z: function (o) {
          var r = -o.utcOffset(),
            t = Math.abs(r),
            n = Math.floor(t / 60),
            e = t % 60;
          return (r <= 0 ? "+" : "-") + K(n, 2, "0") + ":" + K(e, 2, "0");
        },
        m: function o(r, t) {
          if (r.date() < t.date()) return -o(t, r);
          var n = 12 * (t.year() - r.year()) + (t.month() - r.month()),
            e = r.clone().add(n, _),
            i = t - e < 0,
            a = r.clone().add(n + (i ? -1 : 1), _);
          return +(-(n + (t - e) / (i ? e - a : a - e)) || 0);
        },
        a: function (o) {
          return o < 0 ? Math.ceil(o) || 0 : Math.floor(o);
        },
        p: function (o) {
          return (
            { M: _, y: E, w: M, d: l, D: v, h: C, m: D, s: y, ms: T, Q: W }[
              o
            ] ||
            String(o || "")
              .toLowerCase()
              .replace(/s$/, "")
          );
        },
        u: function (o) {
          return o === void 0;
        },
      },
      S = "en",
      $ = {};
    $[S] = j;
    var F = "$isDayjsObject",
      P = function (o) {
        return o instanceof z || !(!o || !o[F]);
      },
      Z = function o(r, t, n) {
        var e;
        if (!r) return S;
        if (typeof r == "string") {
          var i = r.toLowerCase();
          ($[i] && (e = i), t && (($[i] = t), (e = i)));
          var a = r.split("-");
          if (!e && a.length > 1) return o(a[0]);
        } else {
          var u = r.name;
          (($[u] = r), (e = u));
        }
        return (!n && e && (S = e), e || (!n && S));
      },
      f = function (o, r) {
        if (P(o)) return o.clone();
        var t = typeof r == "object" ? r : {};
        return ((t.date = o), (t.args = arguments), new z(t));
      },
      c = I;
    ((c.l = Z),
      (c.i = P),
      (c.w = function (o, r) {
        return f(o, { locale: r.$L, utc: r.$u, x: r.$x, $offset: r.$offset });
      }));
    var z = (function () {
        function o(t) {
          ((this.$L = Z(t.locale, null, !0)),
            this.parse(t),
            (this.$x = this.$x || t.x || {}),
            (this[F] = !0));
        }
        var r = o.prototype;
        return (
          (r.parse = function (t) {
            ((this.$d = (function (n) {
              var e = n.date,
                i = n.utc;
              if (e === null) return new Date(NaN);
              if (c.u(e)) return new Date();
              if (e instanceof Date) return new Date(e);
              if (typeof e == "string" && !/Z$/i.test(e)) {
                var a = e.match(H);
                if (a) {
                  var u = a[2] - 1 || 0,
                    d = (a[7] || "0").substring(0, 3);
                  return i
                    ? new Date(
                        Date.UTC(
                          a[1],
                          u,
                          a[3] || 1,
                          a[4] || 0,
                          a[5] || 0,
                          a[6] || 0,
                          d,
                        ),
                      )
                    : new Date(
                        a[1],
                        u,
                        a[3] || 1,
                        a[4] || 0,
                        a[5] || 0,
                        a[6] || 0,
                        d,
                      );
                }
              }
              return new Date(e);
            })(t)),
              this.init());
          }),
          (r.init = function () {
            var t = this.$d;
            ((this.$y = t.getFullYear()),
              (this.$M = t.getMonth()),
              (this.$D = t.getDate()),
              (this.$W = t.getDay()),
              (this.$H = t.getHours()),
              (this.$m = t.getMinutes()),
              (this.$s = t.getSeconds()),
              (this.$ms = t.getMilliseconds()));
          }),
          (r.$utils = function () {
            return c;
          }),
          (r.isValid = function () {
            return this.$d.toString() !== w;
          }),
          (r.isSame = function (t, n) {
            var e = f(t);
            return this.startOf(n) <= e && e <= this.endOf(n);
          }),
          (r.isAfter = function (t, n) {
            return f(t) < this.startOf(n);
          }),
          (r.isBefore = function (t, n) {
            return this.endOf(n) < f(t);
          }),
          (r.$g = function (t, n, e) {
            return c.u(t) ? this[n] : this.set(e, t);
          }),
          (r.unix = function () {
            return Math.floor(this.valueOf() / 1e3);
          }),
          (r.valueOf = function () {
            return this.$d.getTime();
          }),
          (r.startOf = function (t, n) {
            var e = this,
              i = !!c.u(n) || n,
              a = c.p(t),
              u = function (x, O) {
                var R = c.w(
                  e.$u ? Date.UTC(e.$y, O, x) : new Date(e.$y, O, x),
                  e,
                );
                return i ? R : R.endOf(l);
              },
              d = function (x, O) {
                return c.w(
                  e
                    .toDate()
                    [
                      x
                    ].apply(e.toDate("s"), (i ? [0, 0, 0, 0] : [23, 59, 59, 999]).slice(O)),
                  e,
                );
              },
              h = this.$W,
              p = this.$M,
              A = this.$D,
              U = "set" + (this.$u ? "UTC" : "");
            switch (a) {
              case E:
                return i ? u(1, 0) : u(31, 11);
              case _:
                return i ? u(1, p) : u(0, p + 1);
              case M:
                var L = this.$locale().weekStart || 0,
                  V = (h < L ? h + 7 : h) - L;
                return u(i ? A - V : A + (6 - V), p);
              case l:
              case v:
                return d(U + "Hours", 0);
              case C:
                return d(U + "Minutes", 1);
              case D:
                return d(U + "Seconds", 2);
              case y:
                return d(U + "Milliseconds", 3);
              default:
                return this.clone();
            }
          }),
          (r.endOf = function (t) {
            return this.startOf(t, !1);
          }),
          (r.$set = function (t, n) {
            var e,
              i = c.p(t),
              a = "set" + (this.$u ? "UTC" : ""),
              u = ((e = {}),
              (e[l] = a + "Date"),
              (e[v] = a + "Date"),
              (e[_] = a + "Month"),
              (e[E] = a + "FullYear"),
              (e[C] = a + "Hours"),
              (e[D] = a + "Minutes"),
              (e[y] = a + "Seconds"),
              (e[T] = a + "Milliseconds"),
              e)[i],
              d = i === l ? this.$D + (n - this.$W) : n;
            if (i === _ || i === E) {
              var h = this.clone().set(v, 1);
              (h.$d[u](d),
                h.init(),
                (this.$d = h.set(v, Math.min(this.$D, h.daysInMonth())).$d));
            } else u && this.$d[u](d);
            return (this.init(), this);
          }),
          (r.set = function (t, n) {
            return this.clone().$set(t, n);
          }),
          (r.get = function (t) {
            return this[c.p(t)]();
          }),
          (r.add = function (t, n) {
            var e,
              i = this;
            t = Number(t);
            var a = c.p(n),
              u = function (p) {
                var A = f(i);
                return c.w(A.date(A.date() + Math.round(p * t)), i);
              };
            if (a === _) return this.set(_, this.$M + t);
            if (a === E) return this.set(E, this.$y + t);
            if (a === l) return u(1);
            if (a === M) return u(7);
            var d = ((e = {}), (e[D] = m), (e[C] = b), (e[y] = g), e)[a] || 1,
              h = this.$d.getTime() + t * d;
            return c.w(h, this);
          }),
          (r.subtract = function (t, n) {
            return this.add(-1 * t, n);
          }),
          (r.format = function (t) {
            var n = this,
              e = this.$locale();
            if (!this.isValid()) return e.invalidDate || w;
            var i = t || "YYYY-MM-DDTHH:mm:ssZ",
              a = c.z(this),
              u = this.$H,
              d = this.$m,
              h = this.$M,
              p = e.weekdays,
              A = e.months,
              U = e.meridiem,
              L = function (O, R, G, J) {
                return (O && (O[R] || O(n, i))) || G[R].slice(0, J);
              },
              V = function (O) {
                return c.s(u % 12 || 12, O, "0");
              },
              x =
                U ||
                function (O, R, G) {
                  var J = O < 12 ? "AM" : "PM";
                  return G ? J.toLowerCase() : J;
                };
            return i.replace(Y, function (O, R) {
              return (
                R ||
                (function (G) {
                  switch (G) {
                    case "YY":
                      return String(n.$y).slice(-2);
                    case "YYYY":
                      return c.s(n.$y, 4, "0");
                    case "M":
                      return h + 1;
                    case "MM":
                      return c.s(h + 1, 2, "0");
                    case "MMM":
                      return L(e.monthsShort, h, A, 3);
                    case "MMMM":
                      return L(A, h);
                    case "D":
                      return n.$D;
                    case "DD":
                      return c.s(n.$D, 2, "0");
                    case "d":
                      return String(n.$W);
                    case "dd":
                      return L(e.weekdaysMin, n.$W, p, 2);
                    case "ddd":
                      return L(e.weekdaysShort, n.$W, p, 3);
                    case "dddd":
                      return p[n.$W];
                    case "H":
                      return String(u);
                    case "HH":
                      return c.s(u, 2, "0");
                    case "h":
                      return V(1);
                    case "hh":
                      return V(2);
                    case "a":
                      return x(u, d, !0);
                    case "A":
                      return x(u, d, !1);
                    case "m":
                      return String(d);
                    case "mm":
                      return c.s(d, 2, "0");
                    case "s":
                      return String(n.$s);
                    case "ss":
                      return c.s(n.$s, 2, "0");
                    case "SSS":
                      return c.s(n.$ms, 3, "0");
                    case "Z":
                      return a;
                  }
                  return null;
                })(O) ||
                a.replace(":", "")
              );
            });
          }),
          (r.utcOffset = function () {
            return 15 * -Math.round(this.$d.getTimezoneOffset() / 15);
          }),
          (r.diff = function (t, n, e) {
            var i,
              a = this,
              u = c.p(n),
              d = f(t),
              h = (d.utcOffset() - this.utcOffset()) * m,
              p = this - d,
              A = function () {
                return c.m(a, d);
              };
            switch (u) {
              case E:
                i = A() / 12;
                break;
              case _:
                i = A();
                break;
              case W:
                i = A() / 3;
                break;
              case M:
                i = (p - h) / 6048e5;
                break;
              case l:
                i = (p - h) / 864e5;
                break;
              case C:
                i = p / b;
                break;
              case D:
                i = p / m;
                break;
              case y:
                i = p / g;
                break;
              default:
                i = p;
            }
            return e ? i : c.a(i);
          }),
          (r.daysInMonth = function () {
            return this.endOf(_).$D;
          }),
          (r.$locale = function () {
            return $[this.$L];
          }),
          (r.locale = function (t, n) {
            if (!t) return this.$L;
            var e = this.clone(),
              i = Z(t, n, !0);
            return (i && (e.$L = i), e);
          }),
          (r.clone = function () {
            return c.w(this.$d, this);
          }),
          (r.toDate = function () {
            return new Date(this.valueOf());
          }),
          (r.toJSON = function () {
            return this.isValid() ? this.toISOString() : null;
          }),
          (r.toISOString = function () {
            return this.$d.toISOString();
          }),
          (r.toString = function () {
            return this.$d.toUTCString();
          }),
          o
        );
      })(),
      q = z.prototype;
    return (
      (f.prototype = q),
      [
        ["$ms", T],
        ["$s", y],
        ["$m", D],
        ["$H", C],
        ["$W", l],
        ["$M", _],
        ["$y", E],
        ["$D", v],
      ].forEach(function (o) {
        q[o[1]] = function (r) {
          return this.$g(r, o[0], o[1]);
        };
      }),
      (f.extend = function (o, r) {
        return (o.$i || (o(r, z, f), (o.$i = !0)), f);
      }),
      (f.locale = Z),
      (f.isDayjs = P),
      (f.unix = function (o) {
        return f(1e3 * o);
      }),
      (f.en = $[S]),
      (f.Ls = $),
      (f.p = {}),
      f
    );
  });
})(B);
var rt = B.exports;
const At = X(rt);
var tt = { exports: {} };
(function (s, N) {
  (function (g, m) {
    s.exports = m();
  })(Q, function () {
    return function (g, m, b) {
      g = g || {};
      var T = m.prototype,
        y = {
          future: "in %s",
          past: "%s ago",
          s: "a few seconds",
          m: "a minute",
          mm: "%d minutes",
          h: "an hour",
          hh: "%d hours",
          d: "a day",
          dd: "%d days",
          M: "a month",
          MM: "%d months",
          y: "a year",
          yy: "%d years",
        };
      function D(l, M, _, W) {
        return T.fromToBase(l, M, _, W);
      }
      ((b.en.relativeTime = y),
        (T.fromToBase = function (l, M, _, W, E) {
          for (
            var v,
              w,
              H,
              Y = _.$locale().relativeTime || y,
              j = g.thresholds || [
                { l: "s", r: 44, d: "second" },
                { l: "m", r: 89 },
                { l: "mm", r: 44, d: "minute" },
                { l: "h", r: 89 },
                { l: "hh", r: 21, d: "hour" },
                { l: "d", r: 35 },
                { l: "dd", r: 25, d: "day" },
                { l: "M", r: 45 },
                { l: "MM", r: 10, d: "month" },
                { l: "y", r: 17 },
                { l: "yy", d: "year" },
              ],
              K = j.length,
              I = 0;
            I < K;
            I += 1
          ) {
            var S = j[I];
            S.d && (v = W ? b(l).diff(_, S.d, !0) : _.diff(l, S.d, !0));
            var $ = (g.rounding || Math.round)(Math.abs(v));
            if (((H = v > 0), $ <= S.r || !S.r)) {
              $ <= 1 && I > 0 && (S = j[I - 1]);
              var F = Y[S.l];
              (E && ($ = E("" + $)),
                (w =
                  typeof F == "string" ? F.replace("%d", $) : F($, M, S.l, H)));
              break;
            }
          }
          if (M) return w;
          var P = H ? Y.future : Y.past;
          return typeof P == "function" ? P(w) : P.replace("%s", w);
        }),
        (T.to = function (l, M) {
          return D(l, M, this, !0);
        }),
        (T.from = function (l, M) {
          return D(l, M, this);
        }));
      var C = function (l) {
        return l.$u ? b.utc() : b();
      };
      ((T.toNow = function (l) {
        return this.to(C(this), l);
      }),
        (T.fromNow = function (l) {
          return this.from(C(this), l);
        }));
    };
  });
})(tt);
var nt = tt.exports;
const Et = X(nt);
export {
  ot as A,
  Ot as C,
  et as E,
  ut as I,
  at as O,
  ht as P,
  lt as R,
  pt as S,
  dt as W,
  st as a,
  k as b,
  Q as c,
  it as d,
  At as e,
  rt as f,
  X as g,
  Mt as h,
  _t as i,
  ct as j,
  ft as k,
  gt as l,
  $t as m,
  mt as n,
  St as o,
  Et as r,
};
