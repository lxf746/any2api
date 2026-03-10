var zr = Object.defineProperty;
var Hr = (n, t, e) =>
  t in n
    ? zr(n, t, { enumerable: !0, configurable: !0, writable: !0, value: e })
    : (n[t] = e);
var ee = (n, t, e) => Hr(n, typeof t != "symbol" ? t + "" : t, e);
import {
  h as Yr,
  c as Te,
  g as Le,
  e as se,
  r as Kr,
  b as je,
  A as qr,
} from "./BJ4IMYPr.js";
import { w as k, r as Qr } from "./DOsrtDJS.js";
import { U as ne } from "./DFZQlWS9.js";
import { _ as C } from "./C1FmrZbK.js";
function Gr(n) {
  return {
    all: (n = n || new Map()),
    on: function (t, e) {
      var r = n.get(t);
      r ? r.push(e) : n.set(t, [e]);
    },
    off: function (t, e) {
      var r = n.get(t);
      r && (e ? r.splice(r.indexOf(e) >>> 0, 1) : n.set(t, []));
    },
    emit: function (t, e) {
      var r = n.get(t);
      (r &&
        r.slice().map(function (s) {
          s(e);
        }),
        (r = n.get("*")) &&
          r.slice().map(function (s) {
            s(t, e);
          }));
    },
  };
}
var lt = { exports: {} };
(function (n) {
  n.exports = (function (t) {
    var e = {};
    function r(s) {
      if (e[s]) return e[s].exports;
      var i = (e[s] = { i: s, l: !1, exports: {} });
      return (t[s].call(i.exports, i, i.exports, r), (i.l = !0), i.exports);
    }
    return (
      (r.m = t),
      (r.c = e),
      (r.d = function (s, i, o) {
        r.o(s, i) ||
          Object.defineProperty(s, i, {
            configurable: !1,
            enumerable: !0,
            get: o,
          });
      }),
      (r.n = function (s) {
        var i =
          s && s.__esModule
            ? function () {
                return s.default;
              }
            : function () {
                return s;
              };
        return (r.d(i, "a", i), i);
      }),
      (r.o = function (s, i) {
        return Object.prototype.hasOwnProperty.call(s, i);
      }),
      (r.p = ""),
      r((r.s = 1))
    );
  })([
    function (t, e, r) {
      var s = {
          MOBILE: "mobile",
          TABLET: "tablet",
          SMART_TV: "smarttv",
          CONSOLE: "console",
          WEARABLE: "wearable",
          BROWSER: void 0,
        },
        i = {
          CHROME: "Chrome",
          FIREFOX: "Firefox",
          OPERA: "Opera",
          YANDEX: "Yandex",
          SAFARI: "Safari",
          INTERNET_EXPLORER: "Internet Explorer",
          EDGE: "Edge",
          CHROMIUM: "Chromium",
          IE: "IE",
          MOBILE_SAFARI: "Mobile Safari",
          EDGE_CHROMIUM: "Edge Chromium",
        },
        o = {
          IOS: "iOS",
          ANDROID: "Android",
          WINDOWS_PHONE: "Windows Phone",
          WINDOWS: "Windows",
          MAC_OS: "Mac OS",
        },
        l = {
          isMobile: !1,
          isTablet: !1,
          isBrowser: !1,
          isSmartTV: !1,
          isConsole: !1,
          isWearable: !1,
        };
      t.exports = {
        BROWSER_TYPES: i,
        DEVICE_TYPES: s,
        OS_TYPES: o,
        defaultData: l,
      };
    },
    function (t, e, r) {
      var s = r(2),
        i = r(0),
        o = i.BROWSER_TYPES,
        l = i.OS_TYPES,
        a = i.DEVICE_TYPES,
        c = r(4),
        m = c.checkType,
        v = c.broPayload,
        d = c.mobilePayload,
        w = c.wearPayload,
        E = c.consolePayload,
        p = c.stvPayload,
        x = c.getNavigatorInstance,
        f = c.isIOS13Check,
        h = new s(),
        b = h.getBrowser(),
        L = h.getDevice(),
        W = h.getEngine(),
        u = h.getOS(),
        g = h.getUA(),
        T = o.CHROME,
        S = o.CHROMIUM,
        R = o.IE,
        y = o.INTERNET_EXPLORER,
        P = o.OPERA,
        _ = o.FIREFOX,
        A = o.SAFARI,
        N = o.EDGE,
        U = o.YANDEX,
        $ = o.MOBILE_SAFARI,
        I = a.MOBILE,
        O = a.TABLET,
        F = a.SMART_TV,
        M = a.BROWSER,
        j = a.WEARABLE,
        q = a.CONSOLE,
        V = l.ANDROID,
        H = l.WINDOWS_PHONE,
        Y = l.IOS,
        he = l.WINDOWS,
        pe = l.MAC_OS,
        St = function () {
          return L.type === I;
        },
        Et = function () {
          return L.type === O;
        },
        Ot = function () {
          switch (L.type) {
            case I:
            case O:
              return !0;
            default:
              return !1;
          }
        },
        Ne = function () {
          return u.name === l.WINDOWS && u.version === "10"
            ? typeof g == "string" && g.indexOf("Edg/") !== -1
            : !1;
        },
        Tt = function () {
          return L.type === F;
        },
        Lt = function () {
          return L.type === M;
        },
        Ct = function () {
          return L.type === j;
        },
        kt = function () {
          return L.type === q;
        },
        Rt = function () {
          return u.name === V;
        },
        It = function () {
          return u.name === he;
        },
        Pt = function () {
          return u.name === pe;
        },
        At = function () {
          return u.name === H;
        },
        Ft = function () {
          return u.name === Y;
        },
        Nt = function () {
          return b.name === T;
        },
        Dt = function () {
          return b.name === _;
        },
        Wt = function () {
          return b.name === S;
        },
        De = function () {
          return b.name === N;
        },
        Mt = function () {
          return b.name === U;
        },
        Vt = function () {
          return b.name === A || b.name === $;
        },
        jt = function () {
          return b.name === $;
        },
        Ut = function () {
          return b.name === P;
        },
        $t = function () {
          return b.name === y || b.name === R;
        },
        Bt = function () {
          var G = x(),
            ge = G && G.userAgent.toLowerCase();
          return typeof ge == "string" ? /electron/.test(ge) : !1;
        },
        zt = function () {
          var G = x();
          return (
            G &&
            (/iPad|iPhone|iPod/.test(G.platform) ||
              (G.platform === "MacIntel" && G.maxTouchPoints > 1)) &&
            !window.MSStream
          );
        },
        oe = function () {
          return f("iPad");
        },
        Ht = function () {
          return f("iPhone");
        },
        Yt = function () {
          return f("iPod");
        },
        Kt = function () {
          return b.major;
        },
        qt = function () {
          return b.version;
        },
        Qt = function () {
          return u.version ? u.version : "none";
        },
        Gt = function () {
          return u.name ? u.name : "none";
        },
        Jt = function () {
          return b.name;
        },
        Xt = function () {
          return L.vendor ? L.vendor : "none";
        },
        Zt = function () {
          return L.model ? L.model : "none";
        },
        er = function () {
          return W.name;
        },
        tr = function () {
          return W.version;
        },
        rr = function () {
          return g;
        },
        nr = function () {
          return L.type;
        },
        sr = Tt(),
        ir = kt(),
        or = Ct(),
        ar = jt() || oe(),
        lr = Wt(),
        ur = Ot() || oe(),
        cr = St(),
        fr = Et() || oe(),
        dr = Lt(),
        hr = Rt(),
        pr = At(),
        gr = Ft() || oe(),
        mr = Nt(),
        xr = Dt(),
        br = Vt(),
        vr = Ut(),
        yr = $t(),
        _r = Qt(),
        wr = Gt(),
        Sr = Kt(),
        Er = qt(),
        Or = Jt(),
        Tr = Xt(),
        Lr = Zt(),
        Cr = er(),
        kr = tr(),
        Rr = rr(),
        Ir = De() || Ne(),
        Pr = Mt(),
        Ar = nr(),
        Fr = zt(),
        Nr = oe(),
        Dr = Ht(),
        Wr = Yt(),
        Mr = Bt(),
        Vr = Ne(),
        jr = De(),
        Ur = It(),
        $r = Pt(),
        J = m(L.type);
      function Br() {
        var D = J.isBrowser,
          G = J.isMobile,
          ge = J.isTablet,
          We = J.isSmartTV,
          Me = J.isConsole,
          Ve = J.isWearable;
        if (D) return v(D, b, W, u, g);
        if (We) return p(We, W, u, g);
        if (Me) return E(Me, W, u, g);
        if (G || ge) return d(J, L, u, g);
        if (Ve) return w(Ve, W, u, g);
      }
      t.exports = {
        deviceDetect: Br,
        isSmartTV: sr,
        isConsole: ir,
        isWearable: or,
        isMobileSafari: ar,
        isChromium: lr,
        isMobile: ur,
        isMobileOnly: cr,
        isTablet: fr,
        isBrowser: dr,
        isAndroid: hr,
        isWinPhone: pr,
        isIOS: gr,
        isChrome: mr,
        isFirefox: xr,
        isSafari: br,
        isOpera: vr,
        isIE: yr,
        osVersion: _r,
        osName: wr,
        fullBrowserVersion: Sr,
        browserVersion: Er,
        browserName: Or,
        mobileVendor: Tr,
        mobileModel: Lr,
        engineName: Cr,
        engineVersion: kr,
        getUA: Rr,
        isEdge: Ir,
        isYandex: Pr,
        deviceType: Ar,
        isIOS13: Fr,
        isIPad13: Nr,
        isIPhone13: Dr,
        isIPod13: Wr,
        isElectron: Mr,
        isEdgeChromium: Vr,
        isLegacyEdge: jr,
        isWindows: Ur,
        isMacOs: $r,
      };
    },
    function (t, e, r) {
      var s;
      /*!
       * UAParser.js v0.7.18
       * Lightweight JavaScript-based User-Agent string parser
       * https://github.com/faisalman/ua-parser-js
       *
       * Copyright © 2012-2016 Faisal Salman <fyzlman@gmail.com>
       * Dual licensed under GPLv2 or MIT
       */ (function (i, o) {
        var l = "0.7.18",
          a = "",
          c = "?",
          m = "function",
          v = "undefined",
          d = "object",
          w = "string",
          E = "major",
          p = "model",
          x = "name",
          f = "type",
          h = "vendor",
          b = "version",
          L = "architecture",
          W = "console",
          u = "mobile",
          g = "tablet",
          T = "smarttv",
          S = "wearable",
          R = "embedded",
          y = {
            extend: function (I, O) {
              var F = {};
              for (var M in I)
                O[M] && O[M].length % 2 === 0
                  ? (F[M] = O[M].concat(I[M]))
                  : (F[M] = I[M]);
              return F;
            },
            has: function (I, O) {
              return typeof I == "string"
                ? O.toLowerCase().indexOf(I.toLowerCase()) !== -1
                : !1;
            },
            lowerize: function (I) {
              return I.toLowerCase();
            },
            major: function (I) {
              return typeof I === w
                ? I.replace(/[^\d\.]/g, "").split(".")[0]
                : o;
            },
            trim: function (I) {
              return I.replace(/^[\s\uFEFF\xA0]+|[\s\uFEFF\xA0]+$/g, "");
            },
          },
          P = {
            rgx: function (I, O) {
              for (var F = 0, M, j, q, V, H, Y; F < O.length && !H; ) {
                var he = O[F],
                  pe = O[F + 1];
                for (M = j = 0; M < he.length && !H; )
                  if (((H = he[M++].exec(I)), H))
                    for (q = 0; q < pe.length; q++)
                      ((Y = H[++j]),
                        (V = pe[q]),
                        typeof V === d && V.length > 0
                          ? V.length == 2
                            ? typeof V[1] == m
                              ? (this[V[0]] = V[1].call(this, Y))
                              : (this[V[0]] = V[1])
                            : V.length == 3
                              ? typeof V[1] === m && !(V[1].exec && V[1].test)
                                ? (this[V[0]] = Y
                                    ? V[1].call(this, Y, V[2])
                                    : o)
                                : (this[V[0]] = Y ? Y.replace(V[1], V[2]) : o)
                              : V.length == 4 &&
                                (this[V[0]] = Y
                                  ? V[3].call(this, Y.replace(V[1], V[2]))
                                  : o)
                          : (this[V] = Y || o));
                F += 2;
              }
            },
            str: function (I, O) {
              for (var F in O)
                if (typeof O[F] === d && O[F].length > 0) {
                  for (var M = 0; M < O[F].length; M++)
                    if (y.has(O[F][M], I)) return F === c ? o : F;
                } else if (y.has(O[F], I)) return F === c ? o : F;
              return I;
            },
          },
          _ = {
            browser: {
              oldsafari: {
                version: {
                  "1.0": "/8",
                  1.2: "/1",
                  1.3: "/3",
                  "2.0": "/412",
                  "2.0.2": "/416",
                  "2.0.3": "/417",
                  "2.0.4": "/419",
                  "?": "/",
                },
              },
            },
            device: {
              amazon: { model: { "Fire Phone": ["SD", "KF"] } },
              sprint: {
                model: { "Evo Shift 4G": "7373KT" },
                vendor: { HTC: "APA", Sprint: "Sprint" },
              },
            },
            os: {
              windows: {
                version: {
                  ME: "4.90",
                  "NT 3.11": "NT3.51",
                  "NT 4.0": "NT4.0",
                  2e3: "NT 5.0",
                  XP: ["NT 5.1", "NT 5.2"],
                  Vista: "NT 6.0",
                  7: "NT 6.1",
                  8: "NT 6.2",
                  8.1: "NT 6.3",
                  10: ["NT 6.4", "NT 10.0"],
                  RT: "ARM",
                },
              },
            },
          },
          A = {
            browser: [
              [
                /(opera\smini)\/([\w\.-]+)/i,
                /(opera\s[mobiletab]+).+version\/([\w\.-]+)/i,
                /(opera).+version\/([\w\.]+)/i,
                /(opera)[\/\s]+([\w\.]+)/i,
              ],
              [x, b],
              [/(opios)[\/\s]+([\w\.]+)/i],
              [[x, "Opera Mini"], b],
              [/\s(opr)\/([\w\.]+)/i],
              [[x, "Opera"], b],
              [
                /(kindle)\/([\w\.]+)/i,
                /(lunascape|maxthon|netfront|jasmine|blazer)[\/\s]?([\w\.]*)/i,
                /(avant\s|iemobile|slim|baidu)(?:browser)?[\/\s]?([\w\.]*)/i,
                /(?:ms|\()(ie)\s([\w\.]+)/i,
                /(rekonq)\/([\w\.]*)/i,
                /(chromium|flock|rockmelt|midori|epiphany|silk|skyfire|ovibrowser|bolt|iron|vivaldi|iridium|phantomjs|bowser|quark)\/([\w\.-]+)/i,
              ],
              [x, b],
              [/(trident).+rv[:\s]([\w\.]+).+like\sgecko/i],
              [[x, "IE"], b],
              [/(edge|edgios|edgea)\/((\d+)?[\w\.]+)/i],
              [[x, "Edge"], b],
              [/(yabrowser)\/([\w\.]+)/i],
              [[x, "Yandex"], b],
              [/(puffin)\/([\w\.]+)/i],
              [[x, "Puffin"], b],
              [/((?:[\s\/])uc?\s?browser|(?:juc.+)ucweb)[\/\s]?([\w\.]+)/i],
              [[x, "UCBrowser"], b],
              [/(comodo_dragon)\/([\w\.]+)/i],
              [[x, /_/g, " "], b],
              [/(micromessenger)\/([\w\.]+)/i],
              [[x, "WeChat"], b],
              [/(qqbrowserlite)\/([\w\.]+)/i],
              [x, b],
              [/(QQ)\/([\d\.]+)/i],
              [x, b],
              [/m?(qqbrowser)[\/\s]?([\w\.]+)/i],
              [x, b],
              [/(BIDUBrowser)[\/\s]?([\w\.]+)/i],
              [x, b],
              [/(2345Explorer)[\/\s]?([\w\.]+)/i],
              [x, b],
              [/(MetaSr)[\/\s]?([\w\.]+)/i],
              [x],
              [/(LBBROWSER)/i],
              [x],
              [/xiaomi\/miuibrowser\/([\w\.]+)/i],
              [b, [x, "MIUI Browser"]],
              [/;fbav\/([\w\.]+);/i],
              [b, [x, "Facebook"]],
              [/headlesschrome(?:\/([\w\.]+)|\s)/i],
              [b, [x, "Chrome Headless"]],
              [/\swv\).+(chrome)\/([\w\.]+)/i],
              [[x, /(.+)/, "$1 WebView"], b],
              [/((?:oculus|samsung)browser)\/([\w\.]+)/i],
              [[x, /(.+(?:g|us))(.+)/, "$1 $2"], b],
              [/android.+version\/([\w\.]+)\s+(?:mobile\s?safari|safari)*/i],
              [b, [x, "Android Browser"]],
              [/(chrome|omniweb|arora|[tizenoka]{5}\s?browser)\/v?([\w\.]+)/i],
              [x, b],
              [/(dolfin)\/([\w\.]+)/i],
              [[x, "Dolphin"], b],
              [/((?:android.+)crmo|crios)\/([\w\.]+)/i],
              [[x, "Chrome"], b],
              [/(coast)\/([\w\.]+)/i],
              [[x, "Opera Coast"], b],
              [/fxios\/([\w\.-]+)/i],
              [b, [x, "Firefox"]],
              [/version\/([\w\.]+).+?mobile\/\w+\s(safari)/i],
              [b, [x, "Mobile Safari"]],
              [/version\/([\w\.]+).+?(mobile\s?safari|safari)/i],
              [b, x],
              [
                /webkit.+?(gsa)\/([\w\.]+).+?(mobile\s?safari|safari)(\/[\w\.]+)/i,
              ],
              [[x, "GSA"], b],
              [/webkit.+?(mobile\s?safari|safari)(\/[\w\.]+)/i],
              [x, [b, P.str, _.browser.oldsafari.version]],
              [/(konqueror)\/([\w\.]+)/i, /(webkit|khtml)\/([\w\.]+)/i],
              [x, b],
              [/(navigator|netscape)\/([\w\.-]+)/i],
              [[x, "Netscape"], b],
              [
                /(swiftfox)/i,
                /(icedragon|iceweasel|camino|chimera|fennec|maemo\sbrowser|minimo|conkeror)[\/\s]?([\w\.\+]+)/i,
                /(firefox|seamonkey|k-meleon|icecat|iceape|firebird|phoenix|palemoon|basilisk|waterfox)\/([\w\.-]+)$/i,
                /(mozilla)\/([\w\.]+).+rv\:.+gecko\/\d+/i,
                /(polaris|lynx|dillo|icab|doris|amaya|w3m|netsurf|sleipnir)[\/\s]?([\w\.]+)/i,
                /(links)\s\(([\w\.]+)/i,
                /(gobrowser)\/?([\w\.]*)/i,
                /(ice\s?browser)\/v?([\w\._]+)/i,
                /(mosaic)[\/\s]([\w\.]+)/i,
              ],
              [x, b],
            ],
            cpu: [
              [/(?:(amd|x(?:(?:86|64)[_-])?|wow|win)64)[;\)]/i],
              [[L, "amd64"]],
              [/(ia32(?=;))/i],
              [[L, y.lowerize]],
              [/((?:i[346]|x)86)[;\)]/i],
              [[L, "ia32"]],
              [/windows\s(ce|mobile);\sppc;/i],
              [[L, "arm"]],
              [/((?:ppc|powerpc)(?:64)?)(?:\smac|;|\))/i],
              [[L, /ower/, "", y.lowerize]],
              [/(sun4\w)[;\)]/i],
              [[L, "sparc"]],
              [
                /((?:avr32|ia64(?=;))|68k(?=\))|arm(?:64|(?=v\d+;))|(?=atmel\s)avr|(?:irix|mips|sparc)(?:64)?(?=;)|pa-risc)/i,
              ],
              [[L, y.lowerize]],
            ],
            device: [
              [/\((ipad|playbook);[\w\s\);-]+(rim|apple)/i],
              [p, h, [f, g]],
              [/applecoremedia\/[\w\.]+ \((ipad)/],
              [p, [h, "Apple"], [f, g]],
              [/(apple\s{0,1}tv)/i],
              [
                [p, "Apple TV"],
                [h, "Apple"],
              ],
              [
                /(archos)\s(gamepad2?)/i,
                /(hp).+(touchpad)/i,
                /(hp).+(tablet)/i,
                /(kindle)\/([\w\.]+)/i,
                /\s(nook)[\w\s]+build\/(\w+)/i,
                /(dell)\s(strea[kpr\s\d]*[\dko])/i,
              ],
              [h, p, [f, g]],
              [/(kf[A-z]+)\sbuild\/.+silk\//i],
              [p, [h, "Amazon"], [f, g]],
              [/(sd|kf)[0349hijorstuw]+\sbuild\/.+silk\//i],
              [
                [p, P.str, _.device.amazon.model],
                [h, "Amazon"],
                [f, u],
              ],
              [/\((ip[honed|\s\w*]+);.+(apple)/i],
              [p, h, [f, u]],
              [/\((ip[honed|\s\w*]+);/i],
              [p, [h, "Apple"], [f, u]],
              [
                /(blackberry)[\s-]?(\w+)/i,
                /(blackberry|benq|palm(?=\-)|sonyericsson|acer|asus|dell|meizu|motorola|polytron)[\s_-]?([\w-]*)/i,
                /(hp)\s([\w\s]+\w)/i,
                /(asus)-?(\w+)/i,
              ],
              [h, p, [f, u]],
              [/\(bb10;\s(\w+)/i],
              [p, [h, "BlackBerry"], [f, u]],
              [
                /android.+(transfo[prime\s]{4,10}\s\w+|eeepc|slider\s\w+|nexus 7|padfone)/i,
              ],
              [p, [h, "Asus"], [f, g]],
              [
                /(sony)\s(tablet\s[ps])\sbuild\//i,
                /(sony)?(?:sgp.+)\sbuild\//i,
              ],
              [
                [h, "Sony"],
                [p, "Xperia Tablet"],
                [f, g],
              ],
              [/android.+\s([c-g]\d{4}|so[-l]\w+)\sbuild\//i],
              [p, [h, "Sony"], [f, u]],
              [/\s(ouya)\s/i, /(nintendo)\s([wids3u]+)/i],
              [h, p, [f, W]],
              [/android.+;\s(shield)\sbuild/i],
              [p, [h, "Nvidia"], [f, W]],
              [/(playstation\s[34portablevi]+)/i],
              [p, [h, "Sony"], [f, W]],
              [/(sprint\s(\w+))/i],
              [
                [h, P.str, _.device.sprint.vendor],
                [p, P.str, _.device.sprint.model],
                [f, u],
              ],
              [/(lenovo)\s?(S(?:5000|6000)+(?:[-][\w+]))/i],
              [h, p, [f, g]],
              [
                /(htc)[;_\s-]+([\w\s]+(?=\))|\w+)*/i,
                /(zte)-(\w*)/i,
                /(alcatel|geeksphone|lenovo|nexian|panasonic|(?=;\s)sony)[_\s-]?([\w-]*)/i,
              ],
              [h, [p, /_/g, " "], [f, u]],
              [/(nexus\s9)/i],
              [p, [h, "HTC"], [f, g]],
              [/d\/huawei([\w\s-]+)[;\)]/i, /(nexus\s6p)/i],
              [p, [h, "Huawei"], [f, u]],
              [/(microsoft);\s(lumia[\s\w]+)/i],
              [h, p, [f, u]],
              [/[\s\(;](xbox(?:\sone)?)[\s\);]/i],
              [p, [h, "Microsoft"], [f, W]],
              [/(kin\.[onetw]{3})/i],
              [
                [p, /\./g, " "],
                [h, "Microsoft"],
                [f, u],
              ],
              [
                /\s(milestone|droid(?:[2-4x]|\s(?:bionic|x2|pro|razr))?:?(\s4g)?)[\w\s]+build\//i,
                /mot[\s-]?(\w*)/i,
                /(XT\d{3,4}) build\//i,
                /(nexus\s6)/i,
              ],
              [p, [h, "Motorola"], [f, u]],
              [/android.+\s(mz60\d|xoom[\s2]{0,2})\sbuild\//i],
              [p, [h, "Motorola"], [f, g]],
              [/hbbtv\/\d+\.\d+\.\d+\s+\([\w\s]*;\s*(\w[^;]*);([^;]*)/i],
              [
                [h, y.trim],
                [p, y.trim],
                [f, T],
              ],
              [/hbbtv.+maple;(\d+)/i],
              [
                [p, /^/, "SmartTV"],
                [h, "Samsung"],
                [f, T],
              ],
              [/\(dtv[\);].+(aquos)/i],
              [p, [h, "Sharp"], [f, T]],
              [
                /android.+((sch-i[89]0\d|shw-m380s|gt-p\d{4}|gt-n\d+|sgh-t8[56]9|nexus 10))/i,
                /((SM-T\w+))/i,
              ],
              [[h, "Samsung"], p, [f, g]],
              [/smart-tv.+(samsung)/i],
              [h, [f, T], p],
              [
                /((s[cgp]h-\w+|gt-\w+|galaxy\snexus|sm-\w[\w\d]+))/i,
                /(sam[sung]*)[\s-]*(\w+-?[\w-]*)/i,
                /sec-((sgh\w+))/i,
              ],
              [[h, "Samsung"], p, [f, u]],
              [/sie-(\w*)/i],
              [p, [h, "Siemens"], [f, u]],
              [/(maemo|nokia).*(n900|lumia\s\d+)/i, /(nokia)[\s_-]?([\w-]*)/i],
              [[h, "Nokia"], p, [f, u]],
              [/android\s3\.[\s\w;-]{10}(a\d{3})/i],
              [p, [h, "Acer"], [f, g]],
              [/android.+([vl]k\-?\d{3})\s+build/i],
              [p, [h, "LG"], [f, g]],
              [/android\s3\.[\s\w;-]{10}(lg?)-([06cv9]{3,4})/i],
              [[h, "LG"], p, [f, g]],
              [/(lg) netcast\.tv/i],
              [h, p, [f, T]],
              [
                /(nexus\s[45])/i,
                /lg[e;\s\/-]+(\w*)/i,
                /android.+lg(\-?[\d\w]+)\s+build/i,
              ],
              [p, [h, "LG"], [f, u]],
              [/android.+(ideatab[a-z0-9\-\s]+)/i],
              [p, [h, "Lenovo"], [f, g]],
              [/linux;.+((jolla));/i],
              [h, p, [f, u]],
              [/((pebble))app\/[\d\.]+\s/i],
              [h, p, [f, S]],
              [/android.+;\s(oppo)\s?([\w\s]+)\sbuild/i],
              [h, p, [f, u]],
              [/crkey/i],
              [
                [p, "Chromecast"],
                [h, "Google"],
              ],
              [/android.+;\s(glass)\s\d/i],
              [p, [h, "Google"], [f, S]],
              [/android.+;\s(pixel c)\s/i],
              [p, [h, "Google"], [f, g]],
              [/android.+;\s(pixel xl|pixel)\s/i],
              [p, [h, "Google"], [f, u]],
              [
                /android.+;\s(\w+)\s+build\/hm\1/i,
                /android.+(hm[\s\-_]*note?[\s_]*(?:\d\w)?)\s+build/i,
                /android.+(mi[\s\-_]*(?:one|one[\s_]plus|note lte)?[\s_]*(?:\d?\w?)[\s_]*(?:plus)?)\s+build/i,
                /android.+(redmi[\s\-_]*(?:note)?(?:[\s_]*[\w\s]+))\s+build/i,
              ],
              [
                [p, /_/g, " "],
                [h, "Xiaomi"],
                [f, u],
              ],
              [/android.+(mi[\s\-_]*(?:pad)(?:[\s_]*[\w\s]+))\s+build/i],
              [
                [p, /_/g, " "],
                [h, "Xiaomi"],
                [f, g],
              ],
              [/android.+;\s(m[1-5]\snote)\sbuild/i],
              [p, [h, "Meizu"], [f, g]],
              [
                /android.+a000(1)\s+build/i,
                /android.+oneplus\s(a\d{4})\s+build/i,
              ],
              [p, [h, "OnePlus"], [f, u]],
              [/android.+[;\/]\s*(RCT[\d\w]+)\s+build/i],
              [p, [h, "RCA"], [f, g]],
              [/android.+[;\/\s]+(Venue[\d\s]{2,7})\s+build/i],
              [p, [h, "Dell"], [f, g]],
              [/android.+[;\/]\s*(Q[T|M][\d\w]+)\s+build/i],
              [p, [h, "Verizon"], [f, g]],
              [/android.+[;\/]\s+(Barnes[&\s]+Noble\s+|BN[RT])(V?.*)\s+build/i],
              [[h, "Barnes & Noble"], p, [f, g]],
              [/android.+[;\/]\s+(TM\d{3}.*\b)\s+build/i],
              [p, [h, "NuVision"], [f, g]],
              [/android.+;\s(k88)\sbuild/i],
              [p, [h, "ZTE"], [f, g]],
              [/android.+[;\/]\s*(gen\d{3})\s+build.*49h/i],
              [p, [h, "Swiss"], [f, u]],
              [/android.+[;\/]\s*(zur\d{3})\s+build/i],
              [p, [h, "Swiss"], [f, g]],
              [/android.+[;\/]\s*((Zeki)?TB.*\b)\s+build/i],
              [p, [h, "Zeki"], [f, g]],
              [
                /(android).+[;\/]\s+([YR]\d{2})\s+build/i,
                /android.+[;\/]\s+(Dragon[\-\s]+Touch\s+|DT)(\w{5})\sbuild/i,
              ],
              [[h, "Dragon Touch"], p, [f, g]],
              [/android.+[;\/]\s*(NS-?\w{0,9})\sbuild/i],
              [p, [h, "Insignia"], [f, g]],
              [/android.+[;\/]\s*((NX|Next)-?\w{0,9})\s+build/i],
              [p, [h, "NextBook"], [f, g]],
              [
                /android.+[;\/]\s*(Xtreme\_)?(V(1[045]|2[015]|30|40|60|7[05]|90))\s+build/i,
              ],
              [[h, "Voice"], p, [f, u]],
              [/android.+[;\/]\s*(LVTEL\-)?(V1[12])\s+build/i],
              [[h, "LvTel"], p, [f, u]],
              [/android.+[;\/]\s*(V(100MD|700NA|7011|917G).*\b)\s+build/i],
              [p, [h, "Envizen"], [f, g]],
              [/android.+[;\/]\s*(Le[\s\-]+Pan)[\s\-]+(\w{1,9})\s+build/i],
              [h, p, [f, g]],
              [/android.+[;\/]\s*(Trio[\s\-]*.*)\s+build/i],
              [p, [h, "MachSpeed"], [f, g]],
              [/android.+[;\/]\s*(Trinity)[\-\s]*(T\d{3})\s+build/i],
              [h, p, [f, g]],
              [/android.+[;\/]\s*TU_(1491)\s+build/i],
              [p, [h, "Rotor"], [f, g]],
              [/android.+(KS(.+))\s+build/i],
              [p, [h, "Amazon"], [f, g]],
              [/android.+(Gigaset)[\s\-]+(Q\w{1,9})\s+build/i],
              [h, p, [f, g]],
              [/\s(tablet|tab)[;\/]/i, /\s(mobile)(?:[;\/]|\ssafari)/i],
              [[f, y.lowerize], h, p],
              [/(android[\w\.\s\-]{0,9});.+build/i],
              [p, [h, "Generic"]],
            ],
            engine: [
              [/windows.+\sedge\/([\w\.]+)/i],
              [b, [x, "EdgeHTML"]],
              [
                /(presto)\/([\w\.]+)/i,
                /(webkit|trident|netfront|netsurf|amaya|lynx|w3m)\/([\w\.]+)/i,
                /(khtml|tasman|links)[\/\s]\(?([\w\.]+)/i,
                /(icab)[\/\s]([23]\.[\d\.]+)/i,
              ],
              [x, b],
              [/rv\:([\w\.]{1,9}).+(gecko)/i],
              [b, x],
            ],
            os: [
              [/microsoft\s(windows)\s(vista|xp)/i],
              [x, b],
              [
                /(windows)\snt\s6\.2;\s(arm)/i,
                /(windows\sphone(?:\sos)*)[\s\/]?([\d\.\s\w]*)/i,
                /(windows\smobile|windows)[\s\/]?([ntce\d\.\s]+\w)/i,
              ],
              [x, [b, P.str, _.os.windows.version]],
              [/(win(?=3|9|n)|win\s9x\s)([nt\d\.]+)/i],
              [
                [x, "Windows"],
                [b, P.str, _.os.windows.version],
              ],
              [/\((bb)(10);/i],
              [[x, "BlackBerry"], b],
              [
                /(blackberry)\w*\/?([\w\.]*)/i,
                /(tizen)[\/\s]([\w\.]+)/i,
                /(android|webos|palm\sos|qnx|bada|rim\stablet\sos|meego|contiki)[\/\s-]?([\w\.]*)/i,
                /linux;.+(sailfish);/i,
              ],
              [x, b],
              [/(symbian\s?os|symbos|s60(?=;))[\/\s-]?([\w\.]*)/i],
              [[x, "Symbian"], b],
              [/\((series40);/i],
              [x],
              [/mozilla.+\(mobile;.+gecko.+firefox/i],
              [[x, "Firefox OS"], b],
              [
                /(nintendo|playstation)\s([wids34portablevu]+)/i,
                /(mint)[\/\s\(]?(\w*)/i,
                /(mageia|vectorlinux)[;\s]/i,
                /(joli|[kxln]?ubuntu|debian|suse|opensuse|gentoo|(?=\s)arch|slackware|fedora|mandriva|centos|pclinuxos|redhat|zenwalk|linpus)[\/\s-]?(?!chrom)([\w\.-]*)/i,
                /(hurd|linux)\s?([\w\.]*)/i,
                /(gnu)\s?([\w\.]*)/i,
              ],
              [x, b],
              [/(cros)\s[\w]+\s([\w\.]+\w)/i],
              [[x, "Chromium OS"], b],
              [/(sunos)\s?([\w\.\d]*)/i],
              [[x, "Solaris"], b],
              [/\s([frentopc-]{0,4}bsd|dragonfly)\s?([\w\.]*)/i],
              [x, b],
              [/(haiku)\s(\w+)/i],
              [x, b],
              [
                /cfnetwork\/.+darwin/i,
                /ip[honead]{2,4}(?:.*os\s([\w]+)\slike\smac|;\sopera)/i,
              ],
              [
                [b, /_/g, "."],
                [x, "iOS"],
              ],
              [/(mac\sos\sx)\s?([\w\s\.]*)/i, /(macintosh|mac(?=_powerpc)\s)/i],
              [
                [x, "Mac OS"],
                [b, /_/g, "."],
              ],
              [
                /((?:open)?solaris)[\/\s-]?([\w\.]*)/i,
                /(aix)\s((\d)(?=\.|\)|\s)[\w\.])*/i,
                /(plan\s9|minix|beos|os\/2|amigaos|morphos|risc\sos|openvms)/i,
                /(unix)\s?([\w\.]*)/i,
              ],
              [x, b],
            ],
          },
          N = function (I, O) {
            if (
              (typeof I == "object" && ((O = I), (I = o)), !(this instanceof N))
            )
              return new N(I, O).getResult();
            var F =
                I ||
                (i && i.navigator && i.navigator.userAgent
                  ? i.navigator.userAgent
                  : a),
              M = O ? y.extend(A, O) : A;
            return (
              (this.getBrowser = function () {
                var j = { name: o, version: o };
                return (
                  P.rgx.call(j, F, M.browser),
                  (j.major = y.major(j.version)),
                  j
                );
              }),
              (this.getCPU = function () {
                var j = { architecture: o };
                return (P.rgx.call(j, F, M.cpu), j);
              }),
              (this.getDevice = function () {
                var j = { vendor: o, model: o, type: o };
                return (P.rgx.call(j, F, M.device), j);
              }),
              (this.getEngine = function () {
                var j = { name: o, version: o };
                return (P.rgx.call(j, F, M.engine), j);
              }),
              (this.getOS = function () {
                var j = { name: o, version: o };
                return (P.rgx.call(j, F, M.os), j);
              }),
              (this.getResult = function () {
                return {
                  ua: this.getUA(),
                  browser: this.getBrowser(),
                  engine: this.getEngine(),
                  os: this.getOS(),
                  device: this.getDevice(),
                  cpu: this.getCPU(),
                };
              }),
              (this.getUA = function () {
                return F;
              }),
              (this.setUA = function (j) {
                return ((F = j), this);
              }),
              this
            );
          };
        ((N.VERSION = l),
          (N.BROWSER = { NAME: x, MAJOR: E, VERSION: b }),
          (N.CPU = { ARCHITECTURE: L }),
          (N.DEVICE = {
            MODEL: p,
            VENDOR: h,
            TYPE: f,
            CONSOLE: W,
            MOBILE: u,
            SMARTTV: T,
            TABLET: g,
            WEARABLE: S,
            EMBEDDED: R,
          }),
          (N.ENGINE = { NAME: x, VERSION: b }),
          (N.OS = { NAME: x, VERSION: b }),
          typeof e !== v
            ? (typeof t !== v && t.exports && (e = t.exports = N),
              (e.UAParser = N))
            : r(3)
              ? ((s = function () {
                  return N;
                }.call(e, r, e, t)),
                s !== o && (t.exports = s))
              : i && (i.UAParser = N));
        var U = i && (i.jQuery || i.Zepto);
        if (typeof U !== v) {
          var $ = new N();
          ((U.ua = $.getResult()),
            (U.ua.get = function () {
              return $.getUA();
            }),
            (U.ua.set = function (I) {
              $.setUA(I);
              var O = $.getResult();
              for (var F in O) U.ua[F] = O[F];
            }));
        }
      })(typeof window == "object" ? window : this);
    },
    function (t, e) {
      (function (r) {
        t.exports = r;
      }).call(e, {});
    },
    function (t, e, r) {
      Object.defineProperty(e, "__esModule", { value: !0 });
      var s =
          Object.assign ||
          function (x) {
            for (var f = 1; f < arguments.length; f++) {
              var h = arguments[f];
              for (var b in h)
                Object.prototype.hasOwnProperty.call(h, b) && (x[b] = h[b]);
            }
            return x;
          },
        i = r(0),
        o = i.DEVICE_TYPES,
        l = i.defaultData,
        a = function (f) {
          switch (f) {
            case o.MOBILE:
              return { isMobile: !0 };
            case o.TABLET:
              return { isTablet: !0 };
            case o.SMART_TV:
              return { isSmartTV: !0 };
            case o.CONSOLE:
              return { isConsole: !0 };
            case o.WEARABLE:
              return { isWearable: !0 };
            case o.BROWSER:
              return { isBrowser: !0 };
            default:
              return l;
          }
        },
        c = function (f, h, b, L, W) {
          return {
            isBrowser: f,
            browserMajorVersion: h.major,
            browserFullVersion: h.version,
            browserName: h.name,
            engineName: b.name || !1,
            engineVersion: b.version,
            osName: L.name,
            osVersion: L.version,
            userAgent: W,
          };
        },
        m = function (f, h, b, L) {
          return s({}, f, {
            vendor: h.vendor,
            model: h.model,
            os: b.name,
            osVersion: b.version,
            ua: L,
          });
        },
        v = function (f, h, b, L) {
          return {
            isSmartTV: f,
            engineName: h.name,
            engineVersion: h.version,
            osName: b.name,
            osVersion: b.version,
            userAgent: L,
          };
        },
        d = function (f, h, b, L) {
          return {
            isConsole: f,
            engineName: h.name,
            engineVersion: h.version,
            osName: b.name,
            osVersion: b.version,
            userAgent: L,
          };
        },
        w = function (f, h, b, L) {
          return {
            isWearable: f,
            engineName: h.name,
            engineVersion: h.version,
            osName: b.name,
            osVersion: b.version,
            userAgent: L,
          };
        },
        E = (e.getNavigatorInstance = function () {
          return typeof window < "u" && (window.navigator || navigator)
            ? window.navigator || navigator
            : !1;
        }),
        p = (e.isIOS13Check = function (f) {
          var h = E();
          return (
            h &&
            h.platform &&
            (h.platform.indexOf(f) !== -1 ||
              (h.platform === "MacIntel" &&
                h.maxTouchPoints > 1 &&
                !window.MSStream))
          );
        });
      t.exports = {
        checkType: a,
        broPayload: c,
        mobilePayload: m,
        stvPayload: v,
        consolePayload: d,
        wearPayload: w,
        getNavigatorInstance: E,
        isIOS13Check: p,
      };
    },
  ]);
})(lt);
var Z = lt.exports;
let me;
const Jr = new Uint8Array(16);
function Xr() {
  if (
    !me &&
    ((me =
      typeof crypto < "u" &&
      crypto.getRandomValues &&
      crypto.getRandomValues.bind(crypto)),
    !me)
  )
    throw new Error(
      "crypto.getRandomValues() not supported. See https://github.com/uuidjs/uuid#getrandomvalues-not-supported",
    );
  return me(Jr);
}
const z = [];
for (let n = 0; n < 256; ++n) z.push((n + 256).toString(16).slice(1));
function Zr(n, t = 0) {
  return (
    z[n[t + 0]] +
    z[n[t + 1]] +
    z[n[t + 2]] +
    z[n[t + 3]] +
    "-" +
    z[n[t + 4]] +
    z[n[t + 5]] +
    "-" +
    z[n[t + 6]] +
    z[n[t + 7]] +
    "-" +
    z[n[t + 8]] +
    z[n[t + 9]] +
    "-" +
    z[n[t + 10]] +
    z[n[t + 11]] +
    z[n[t + 12]] +
    z[n[t + 13]] +
    z[n[t + 14]] +
    z[n[t + 15]]
  );
}
const en =
    typeof crypto < "u" && crypto.randomUUID && crypto.randomUUID.bind(crypto),
  Ue = { randomUUID: en };
function ut(n, t, e) {
  if (Ue.randomUUID && !t && !n) return Ue.randomUUID();
  n = n || {};
  const r = n.random || (n.rng || Xr)();
  return ((r[6] = (r[6] & 15) | 64), (r[8] = (r[8] & 63) | 128), Zr(r));
}
var ct = { exports: {} };
const tn = {},
  rn = Object.freeze(
    Object.defineProperty(
      { __proto__: null, default: tn },
      Symbol.toStringTag,
      { value: "Module" },
    ),
  ),
  $e = Yr(rn);
/**
 * [js-sha256]{@link https://github.com/emn178/js-sha256}
 *
 * @version 0.10.1
 * @author Chen, Yi-Cyuan [emn178@gmail.com]
 * @copyright Chen, Yi-Cyuan 2014-2023
 * @license MIT
 */ (function (n) {
  (function () {
    var t = "input is invalid type",
      e = typeof window == "object",
      r = e ? window : {};
    r.JS_SHA256_NO_WINDOW && (e = !1);
    var s = !e && typeof self == "object",
      i =
        !r.JS_SHA256_NO_NODE_JS &&
        typeof process == "object" &&
        process.versions &&
        process.versions.node;
    i ? (r = Te) : s && (r = self);
    var o = !r.JS_SHA256_NO_COMMON_JS && !0 && n.exports,
      l = !r.JS_SHA256_NO_ARRAY_BUFFER && typeof ArrayBuffer < "u",
      a = "0123456789abcdef".split(""),
      c = [-2147483648, 8388608, 32768, 128],
      m = [24, 16, 8, 0],
      v = [
        1116352408, 1899447441, 3049323471, 3921009573, 961987163, 1508970993,
        2453635748, 2870763221, 3624381080, 310598401, 607225278, 1426881987,
        1925078388, 2162078206, 2614888103, 3248222580, 3835390401, 4022224774,
        264347078, 604807628, 770255983, 1249150122, 1555081692, 1996064986,
        2554220882, 2821834349, 2952996808, 3210313671, 3336571891, 3584528711,
        113926993, 338241895, 666307205, 773529912, 1294757372, 1396182291,
        1695183700, 1986661051, 2177026350, 2456956037, 2730485921, 2820302411,
        3259730800, 3345764771, 3516065817, 3600352804, 4094571909, 275423344,
        430227734, 506948616, 659060556, 883997877, 958139571, 1322822218,
        1537002063, 1747873779, 1955562222, 2024104815, 2227730452, 2361852424,
        2428436474, 2756734187, 3204031479, 3329325298,
      ],
      d = ["hex", "array", "digest", "arrayBuffer"],
      w = [];
    ((r.JS_SHA256_NO_NODE_JS || !Array.isArray) &&
      (Array.isArray = function (u) {
        return Object.prototype.toString.call(u) === "[object Array]";
      }),
      l &&
        (r.JS_SHA256_NO_ARRAY_BUFFER_IS_VIEW || !ArrayBuffer.isView) &&
        (ArrayBuffer.isView = function (u) {
          return (
            typeof u == "object" &&
            u.buffer &&
            u.buffer.constructor === ArrayBuffer
          );
        }));
    var E = function (u, g) {
        return function (T) {
          return new b(g, !0).update(T)[u]();
        };
      },
      p = function (u) {
        var g = E("hex", u);
        (i && (g = x(g, u)),
          (g.create = function () {
            return new b(u);
          }),
          (g.update = function (R) {
            return g.create().update(R);
          }));
        for (var T = 0; T < d.length; ++T) {
          var S = d[T];
          g[S] = E(S, u);
        }
        return g;
      },
      x = function (u, g) {
        var T = $e,
          S = $e.Buffer,
          R = g ? "sha224" : "sha256",
          y;
        S.from && !r.JS_SHA256_NO_BUFFER_FROM
          ? (y = S.from)
          : (y = function (_) {
              return new S(_);
            });
        var P = function (_) {
          if (typeof _ == "string")
            return T.createHash(R).update(_, "utf8").digest("hex");
          if (_ == null) throw new Error(t);
          return (
            _.constructor === ArrayBuffer && (_ = new Uint8Array(_)),
            Array.isArray(_) || ArrayBuffer.isView(_) || _.constructor === S
              ? T.createHash(R).update(y(_)).digest("hex")
              : u(_)
          );
        };
        return P;
      },
      f = function (u, g) {
        return function (T, S) {
          return new L(T, g, !0).update(S)[u]();
        };
      },
      h = function (u) {
        var g = f("hex", u);
        ((g.create = function (R) {
          return new L(R, u);
        }),
          (g.update = function (R, y) {
            return g.create(R).update(y);
          }));
        for (var T = 0; T < d.length; ++T) {
          var S = d[T];
          g[S] = f(S, u);
        }
        return g;
      };
    function b(u, g) {
      (g
        ? ((w[0] =
            w[16] =
            w[1] =
            w[2] =
            w[3] =
            w[4] =
            w[5] =
            w[6] =
            w[7] =
            w[8] =
            w[9] =
            w[10] =
            w[11] =
            w[12] =
            w[13] =
            w[14] =
            w[15] =
              0),
          (this.blocks = w))
        : (this.blocks = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]),
        u
          ? ((this.h0 = 3238371032),
            (this.h1 = 914150663),
            (this.h2 = 812702999),
            (this.h3 = 4144912697),
            (this.h4 = 4290775857),
            (this.h5 = 1750603025),
            (this.h6 = 1694076839),
            (this.h7 = 3204075428))
          : ((this.h0 = 1779033703),
            (this.h1 = 3144134277),
            (this.h2 = 1013904242),
            (this.h3 = 2773480762),
            (this.h4 = 1359893119),
            (this.h5 = 2600822924),
            (this.h6 = 528734635),
            (this.h7 = 1541459225)),
        (this.block = this.start = this.bytes = this.hBytes = 0),
        (this.finalized = this.hashed = !1),
        (this.first = !0),
        (this.is224 = u));
    }
    ((b.prototype.update = function (u) {
      if (!this.finalized) {
        var g,
          T = typeof u;
        if (T !== "string") {
          if (T === "object") {
            if (u === null) throw new Error(t);
            if (l && u.constructor === ArrayBuffer) u = new Uint8Array(u);
            else if (!Array.isArray(u) && (!l || !ArrayBuffer.isView(u)))
              throw new Error(t);
          } else throw new Error(t);
          g = !0;
        }
        for (var S, R = 0, y, P = u.length, _ = this.blocks; R < P; ) {
          if (
            (this.hashed &&
              ((this.hashed = !1),
              (_[0] = this.block),
              (_[16] =
                _[1] =
                _[2] =
                _[3] =
                _[4] =
                _[5] =
                _[6] =
                _[7] =
                _[8] =
                _[9] =
                _[10] =
                _[11] =
                _[12] =
                _[13] =
                _[14] =
                _[15] =
                  0)),
            g)
          )
            for (y = this.start; R < P && y < 64; ++R)
              _[y >> 2] |= u[R] << m[y++ & 3];
          else
            for (y = this.start; R < P && y < 64; ++R)
              ((S = u.charCodeAt(R)),
                S < 128
                  ? (_[y >> 2] |= S << m[y++ & 3])
                  : S < 2048
                    ? ((_[y >> 2] |= (192 | (S >> 6)) << m[y++ & 3]),
                      (_[y >> 2] |= (128 | (S & 63)) << m[y++ & 3]))
                    : S < 55296 || S >= 57344
                      ? ((_[y >> 2] |= (224 | (S >> 12)) << m[y++ & 3]),
                        (_[y >> 2] |= (128 | ((S >> 6) & 63)) << m[y++ & 3]),
                        (_[y >> 2] |= (128 | (S & 63)) << m[y++ & 3]))
                      : ((S =
                          65536 +
                          (((S & 1023) << 10) | (u.charCodeAt(++R) & 1023))),
                        (_[y >> 2] |= (240 | (S >> 18)) << m[y++ & 3]),
                        (_[y >> 2] |= (128 | ((S >> 12) & 63)) << m[y++ & 3]),
                        (_[y >> 2] |= (128 | ((S >> 6) & 63)) << m[y++ & 3]),
                        (_[y >> 2] |= (128 | (S & 63)) << m[y++ & 3])));
          ((this.lastByteIndex = y),
            (this.bytes += y - this.start),
            y >= 64
              ? ((this.block = _[16]),
                (this.start = y - 64),
                this.hash(),
                (this.hashed = !0))
              : (this.start = y));
        }
        return (
          this.bytes > 4294967295 &&
            ((this.hBytes += (this.bytes / 4294967296) << 0),
            (this.bytes = this.bytes % 4294967296)),
          this
        );
      }
    }),
      (b.prototype.finalize = function () {
        if (!this.finalized) {
          this.finalized = !0;
          var u = this.blocks,
            g = this.lastByteIndex;
          ((u[16] = this.block),
            (u[g >> 2] |= c[g & 3]),
            (this.block = u[16]),
            g >= 56 &&
              (this.hashed || this.hash(),
              (u[0] = this.block),
              (u[16] =
                u[1] =
                u[2] =
                u[3] =
                u[4] =
                u[5] =
                u[6] =
                u[7] =
                u[8] =
                u[9] =
                u[10] =
                u[11] =
                u[12] =
                u[13] =
                u[14] =
                u[15] =
                  0)),
            (u[14] = (this.hBytes << 3) | (this.bytes >>> 29)),
            (u[15] = this.bytes << 3),
            this.hash());
        }
      }),
      (b.prototype.hash = function () {
        var u = this.h0,
          g = this.h1,
          T = this.h2,
          S = this.h3,
          R = this.h4,
          y = this.h5,
          P = this.h6,
          _ = this.h7,
          A = this.blocks,
          N,
          U,
          $,
          I,
          O,
          F,
          M,
          j,
          q,
          V,
          H;
        for (N = 16; N < 64; ++N)
          ((O = A[N - 15]),
            (U =
              ((O >>> 7) | (O << 25)) ^ ((O >>> 18) | (O << 14)) ^ (O >>> 3)),
            (O = A[N - 2]),
            ($ =
              ((O >>> 17) | (O << 15)) ^ ((O >>> 19) | (O << 13)) ^ (O >>> 10)),
            (A[N] = (A[N - 16] + U + A[N - 7] + $) << 0));
        for (H = g & T, N = 0; N < 64; N += 4)
          (this.first
            ? (this.is224
                ? ((j = 300032),
                  (O = A[0] - 1413257819),
                  (_ = (O - 150054599) << 0),
                  (S = (O + 24177077) << 0))
                : ((j = 704751109),
                  (O = A[0] - 210244248),
                  (_ = (O - 1521486534) << 0),
                  (S = (O + 143694565) << 0)),
              (this.first = !1))
            : ((U =
                ((u >>> 2) | (u << 30)) ^
                ((u >>> 13) | (u << 19)) ^
                ((u >>> 22) | (u << 10))),
              ($ =
                ((R >>> 6) | (R << 26)) ^
                ((R >>> 11) | (R << 21)) ^
                ((R >>> 25) | (R << 7))),
              (j = u & g),
              (I = j ^ (u & T) ^ H),
              (M = (R & y) ^ (~R & P)),
              (O = _ + $ + M + v[N] + A[N]),
              (F = U + I),
              (_ = (S + O) << 0),
              (S = (O + F) << 0)),
            (U =
              ((S >>> 2) | (S << 30)) ^
              ((S >>> 13) | (S << 19)) ^
              ((S >>> 22) | (S << 10))),
            ($ =
              ((_ >>> 6) | (_ << 26)) ^
              ((_ >>> 11) | (_ << 21)) ^
              ((_ >>> 25) | (_ << 7))),
            (q = S & u),
            (I = q ^ (S & g) ^ j),
            (M = (_ & R) ^ (~_ & y)),
            (O = P + $ + M + v[N + 1] + A[N + 1]),
            (F = U + I),
            (P = (T + O) << 0),
            (T = (O + F) << 0),
            (U =
              ((T >>> 2) | (T << 30)) ^
              ((T >>> 13) | (T << 19)) ^
              ((T >>> 22) | (T << 10))),
            ($ =
              ((P >>> 6) | (P << 26)) ^
              ((P >>> 11) | (P << 21)) ^
              ((P >>> 25) | (P << 7))),
            (V = T & S),
            (I = V ^ (T & u) ^ q),
            (M = (P & _) ^ (~P & R)),
            (O = y + $ + M + v[N + 2] + A[N + 2]),
            (F = U + I),
            (y = (g + O) << 0),
            (g = (O + F) << 0),
            (U =
              ((g >>> 2) | (g << 30)) ^
              ((g >>> 13) | (g << 19)) ^
              ((g >>> 22) | (g << 10))),
            ($ =
              ((y >>> 6) | (y << 26)) ^
              ((y >>> 11) | (y << 21)) ^
              ((y >>> 25) | (y << 7))),
            (H = g & T),
            (I = H ^ (g & S) ^ V),
            (M = (y & P) ^ (~y & _)),
            (O = R + $ + M + v[N + 3] + A[N + 3]),
            (F = U + I),
            (R = (u + O) << 0),
            (u = (O + F) << 0),
            (this.chromeBugWorkAround = !0));
        ((this.h0 = (this.h0 + u) << 0),
          (this.h1 = (this.h1 + g) << 0),
          (this.h2 = (this.h2 + T) << 0),
          (this.h3 = (this.h3 + S) << 0),
          (this.h4 = (this.h4 + R) << 0),
          (this.h5 = (this.h5 + y) << 0),
          (this.h6 = (this.h6 + P) << 0),
          (this.h7 = (this.h7 + _) << 0));
      }),
      (b.prototype.hex = function () {
        this.finalize();
        var u = this.h0,
          g = this.h1,
          T = this.h2,
          S = this.h3,
          R = this.h4,
          y = this.h5,
          P = this.h6,
          _ = this.h7,
          A =
            a[(u >> 28) & 15] +
            a[(u >> 24) & 15] +
            a[(u >> 20) & 15] +
            a[(u >> 16) & 15] +
            a[(u >> 12) & 15] +
            a[(u >> 8) & 15] +
            a[(u >> 4) & 15] +
            a[u & 15] +
            a[(g >> 28) & 15] +
            a[(g >> 24) & 15] +
            a[(g >> 20) & 15] +
            a[(g >> 16) & 15] +
            a[(g >> 12) & 15] +
            a[(g >> 8) & 15] +
            a[(g >> 4) & 15] +
            a[g & 15] +
            a[(T >> 28) & 15] +
            a[(T >> 24) & 15] +
            a[(T >> 20) & 15] +
            a[(T >> 16) & 15] +
            a[(T >> 12) & 15] +
            a[(T >> 8) & 15] +
            a[(T >> 4) & 15] +
            a[T & 15] +
            a[(S >> 28) & 15] +
            a[(S >> 24) & 15] +
            a[(S >> 20) & 15] +
            a[(S >> 16) & 15] +
            a[(S >> 12) & 15] +
            a[(S >> 8) & 15] +
            a[(S >> 4) & 15] +
            a[S & 15] +
            a[(R >> 28) & 15] +
            a[(R >> 24) & 15] +
            a[(R >> 20) & 15] +
            a[(R >> 16) & 15] +
            a[(R >> 12) & 15] +
            a[(R >> 8) & 15] +
            a[(R >> 4) & 15] +
            a[R & 15] +
            a[(y >> 28) & 15] +
            a[(y >> 24) & 15] +
            a[(y >> 20) & 15] +
            a[(y >> 16) & 15] +
            a[(y >> 12) & 15] +
            a[(y >> 8) & 15] +
            a[(y >> 4) & 15] +
            a[y & 15] +
            a[(P >> 28) & 15] +
            a[(P >> 24) & 15] +
            a[(P >> 20) & 15] +
            a[(P >> 16) & 15] +
            a[(P >> 12) & 15] +
            a[(P >> 8) & 15] +
            a[(P >> 4) & 15] +
            a[P & 15];
        return (
          this.is224 ||
            (A +=
              a[(_ >> 28) & 15] +
              a[(_ >> 24) & 15] +
              a[(_ >> 20) & 15] +
              a[(_ >> 16) & 15] +
              a[(_ >> 12) & 15] +
              a[(_ >> 8) & 15] +
              a[(_ >> 4) & 15] +
              a[_ & 15]),
          A
        );
      }),
      (b.prototype.toString = b.prototype.hex),
      (b.prototype.digest = function () {
        this.finalize();
        var u = this.h0,
          g = this.h1,
          T = this.h2,
          S = this.h3,
          R = this.h4,
          y = this.h5,
          P = this.h6,
          _ = this.h7,
          A = [
            (u >> 24) & 255,
            (u >> 16) & 255,
            (u >> 8) & 255,
            u & 255,
            (g >> 24) & 255,
            (g >> 16) & 255,
            (g >> 8) & 255,
            g & 255,
            (T >> 24) & 255,
            (T >> 16) & 255,
            (T >> 8) & 255,
            T & 255,
            (S >> 24) & 255,
            (S >> 16) & 255,
            (S >> 8) & 255,
            S & 255,
            (R >> 24) & 255,
            (R >> 16) & 255,
            (R >> 8) & 255,
            R & 255,
            (y >> 24) & 255,
            (y >> 16) & 255,
            (y >> 8) & 255,
            y & 255,
            (P >> 24) & 255,
            (P >> 16) & 255,
            (P >> 8) & 255,
            P & 255,
          ];
        return (
          this.is224 ||
            A.push((_ >> 24) & 255, (_ >> 16) & 255, (_ >> 8) & 255, _ & 255),
          A
        );
      }),
      (b.prototype.array = b.prototype.digest),
      (b.prototype.arrayBuffer = function () {
        this.finalize();
        var u = new ArrayBuffer(this.is224 ? 28 : 32),
          g = new DataView(u);
        return (
          g.setUint32(0, this.h0),
          g.setUint32(4, this.h1),
          g.setUint32(8, this.h2),
          g.setUint32(12, this.h3),
          g.setUint32(16, this.h4),
          g.setUint32(20, this.h5),
          g.setUint32(24, this.h6),
          this.is224 || g.setUint32(28, this.h7),
          u
        );
      }));
    function L(u, g, T) {
      var S,
        R = typeof u;
      if (R === "string") {
        var y = [],
          P = u.length,
          _ = 0,
          A;
        for (S = 0; S < P; ++S)
          ((A = u.charCodeAt(S)),
            A < 128
              ? (y[_++] = A)
              : A < 2048
                ? ((y[_++] = 192 | (A >> 6)), (y[_++] = 128 | (A & 63)))
                : A < 55296 || A >= 57344
                  ? ((y[_++] = 224 | (A >> 12)),
                    (y[_++] = 128 | ((A >> 6) & 63)),
                    (y[_++] = 128 | (A & 63)))
                  : ((A =
                      65536 +
                      (((A & 1023) << 10) | (u.charCodeAt(++S) & 1023))),
                    (y[_++] = 240 | (A >> 18)),
                    (y[_++] = 128 | ((A >> 12) & 63)),
                    (y[_++] = 128 | ((A >> 6) & 63)),
                    (y[_++] = 128 | (A & 63))));
        u = y;
      } else if (R === "object") {
        if (u === null) throw new Error(t);
        if (l && u.constructor === ArrayBuffer) u = new Uint8Array(u);
        else if (!Array.isArray(u) && (!l || !ArrayBuffer.isView(u)))
          throw new Error(t);
      } else throw new Error(t);
      u.length > 64 && (u = new b(g, !0).update(u).array());
      var N = [],
        U = [];
      for (S = 0; S < 64; ++S) {
        var $ = u[S] || 0;
        ((N[S] = 92 ^ $), (U[S] = 54 ^ $));
      }
      (b.call(this, g, T),
        this.update(U),
        (this.oKeyPad = N),
        (this.inner = !0),
        (this.sharedMemory = T));
    }
    ((L.prototype = new b()),
      (L.prototype.finalize = function () {
        if ((b.prototype.finalize.call(this), this.inner)) {
          this.inner = !1;
          var u = this.array();
          (b.call(this, this.is224, this.sharedMemory),
            this.update(this.oKeyPad),
            this.update(u),
            b.prototype.finalize.call(this));
        }
      }));
    var W = p();
    ((W.sha256 = W),
      (W.sha224 = p(!0)),
      (W.sha256.hmac = h()),
      (W.sha224.hmac = h(!0)),
      o ? (n.exports = W) : ((r.sha256 = W.sha256), (r.sha224 = W.sha224)));
  })();
})(ct);
var nn = ct.exports;
const Be = Le(nn);
var ft = { exports: {} };
(function (n, t) {
  (function (e, r) {
    n.exports = r();
  })(Te, function () {
    return function (e, r, s) {
      r.prototype.isToday = function () {
        var i = "YYYY-MM-DD",
          o = s();
        return this.format(i) === o.format(i);
      };
    };
  });
})(ft);
var sn = ft.exports;
const on = Le(sn);
var dt = { exports: {} };
(function (n, t) {
  (function (e, r) {
    n.exports = r();
  })(Te, function () {
    return function (e, r, s) {
      r.prototype.isYesterday = function () {
        var i = "YYYY-MM-DD",
          o = s().subtract(1, "day");
        return this.format(i) === o.format(i);
      };
    };
  });
})(dt);
var an = dt.exports;
const ln = Le(an);
var ht = { exports: {} };
(function (n, t) {
  (function (e, r) {
    n.exports = r();
  })(Te, function () {
    var e = {
      LTS: "h:mm:ss A",
      LT: "h:mm A",
      L: "MM/DD/YYYY",
      LL: "MMMM D, YYYY",
      LLL: "MMMM D, YYYY h:mm A",
      LLLL: "dddd, MMMM D, YYYY h:mm A",
    };
    return function (r, s, i) {
      var o = s.prototype,
        l = o.format;
      ((i.en.formats = e),
        (o.format = function (a) {
          a === void 0 && (a = "YYYY-MM-DDTHH:mm:ssZ");
          var c = this.$locale().formats,
            m = (function (v, d) {
              return v.replace(
                /(\[[^\]]+])|(LTS?|l{1,4}|L{1,4})/g,
                function (w, E, p) {
                  var x = p && p.toUpperCase();
                  return (
                    E ||
                    d[p] ||
                    e[p] ||
                    d[x].replace(
                      /(\[[^\]]+])|(MMMM|MM|DD|dddd)/g,
                      function (f, h, b) {
                        return h || b.slice(1);
                      },
                    )
                  );
                },
              );
            })(a, c === void 0 ? {} : c);
          return l.call(this, m);
        }));
    };
  });
})(ht);
var un = ht.exports;
const cn = Le(un),
  fn = (n, t, e) => {
    const r = n[t];
    return r
      ? typeof r == "function"
        ? r()
        : Promise.resolve(r)
      : new Promise((s, i) => {
          (typeof queueMicrotask == "function" ? queueMicrotask : setTimeout)(
            i.bind(
              null,
              new Error(
                "Unknown variable dynamic import: " +
                  t +
                  (t.split("/").length !== e
                    ? ". Note that variables only represent file names one level deep."
                    : ""),
              ),
            ),
          );
        });
  },
  dn = {
    type: "logger",
    log(n) {
      this.output("log", n);
    },
    warn(n) {
      this.output("warn", n);
    },
    error(n) {
      this.output("error", n);
    },
    output(n, t) {
      console && console[n] && console[n].apply(console, t);
    },
  };
class ve {
  constructor(t) {
    let e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
    this.init(t, e);
  }
  init(t) {
    let e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
    ((this.prefix = e.prefix || "i18next:"),
      (this.logger = t || dn),
      (this.options = e),
      (this.debug = e.debug));
  }
  log() {
    for (var t = arguments.length, e = new Array(t), r = 0; r < t; r++)
      e[r] = arguments[r];
    return this.forward(e, "log", "", !0);
  }
  warn() {
    for (var t = arguments.length, e = new Array(t), r = 0; r < t; r++)
      e[r] = arguments[r];
    return this.forward(e, "warn", "", !0);
  }
  error() {
    for (var t = arguments.length, e = new Array(t), r = 0; r < t; r++)
      e[r] = arguments[r];
    return this.forward(e, "error", "");
  }
  deprecate() {
    for (var t = arguments.length, e = new Array(t), r = 0; r < t; r++)
      e[r] = arguments[r];
    return this.forward(e, "warn", "WARNING DEPRECATED: ", !0);
  }
  forward(t, e, r, s) {
    return s && !this.debug
      ? null
      : (typeof t[0] == "string" && (t[0] = `${r}${this.prefix} ${t[0]}`),
        this.logger[e](t));
  }
  create(t) {
    return new ve(this.logger, {
      prefix: `${this.prefix}:${t}:`,
      ...this.options,
    });
  }
  clone(t) {
    return (
      (t = t || this.options),
      (t.prefix = t.prefix || this.prefix),
      new ve(this.logger, t)
    );
  }
}
var Q = new ve();
class Ce {
  constructor() {
    this.observers = {};
  }
  on(t, e) {
    return (
      t.split(" ").forEach((r) => {
        this.observers[r] || (this.observers[r] = new Map());
        const s = this.observers[r].get(e) || 0;
        this.observers[r].set(e, s + 1);
      }),
      this
    );
  }
  off(t, e) {
    if (this.observers[t]) {
      if (!e) {
        delete this.observers[t];
        return;
      }
      this.observers[t].delete(e);
    }
  }
  emit(t) {
    for (
      var e = arguments.length, r = new Array(e > 1 ? e - 1 : 0), s = 1;
      s < e;
      s++
    )
      r[s - 1] = arguments[s];
    (this.observers[t] &&
      Array.from(this.observers[t].entries()).forEach((o) => {
        let [l, a] = o;
        for (let c = 0; c < a; c++) l(...r);
      }),
      this.observers["*"] &&
        Array.from(this.observers["*"].entries()).forEach((o) => {
          let [l, a] = o;
          for (let c = 0; c < a; c++) l.apply(l, [t, ...r]);
        }));
  }
}
function ae() {
  let n, t;
  const e = new Promise((r, s) => {
    ((n = r), (t = s));
  });
  return ((e.resolve = n), (e.reject = t), e);
}
function ze(n) {
  return n == null ? "" : "" + n;
}
function hn(n, t, e) {
  n.forEach((r) => {
    t[r] && (e[r] = t[r]);
  });
}
const pn = /###/g;
function ce(n, t, e) {
  function r(l) {
    return l && l.indexOf("###") > -1 ? l.replace(pn, ".") : l;
  }
  function s() {
    return !n || typeof n == "string";
  }
  const i = typeof t != "string" ? t : t.split(".");
  let o = 0;
  for (; o < i.length - 1; ) {
    if (s()) return {};
    const l = r(i[o]);
    (!n[l] && e && (n[l] = new e()),
      Object.prototype.hasOwnProperty.call(n, l) ? (n = n[l]) : (n = {}),
      ++o);
  }
  return s() ? {} : { obj: n, k: r(i[o]) };
}
function He(n, t, e) {
  const { obj: r, k: s } = ce(n, t, Object);
  if (r !== void 0 || t.length === 1) {
    r[s] = e;
    return;
  }
  let i = t[t.length - 1],
    o = t.slice(0, t.length - 1),
    l = ce(n, o, Object);
  for (; l.obj === void 0 && o.length; )
    ((i = `${o[o.length - 1]}.${i}`),
      (o = o.slice(0, o.length - 1)),
      (l = ce(n, o, Object)),
      l && l.obj && typeof l.obj[`${l.k}.${i}`] < "u" && (l.obj = void 0));
  l.obj[`${l.k}.${i}`] = e;
}
function gn(n, t, e, r) {
  const { obj: s, k: i } = ce(n, t, Object);
  ((s[i] = s[i] || []), s[i].push(e));
}
function ye(n, t) {
  const { obj: e, k: r } = ce(n, t);
  if (e) return e[r];
}
function mn(n, t, e) {
  const r = ye(n, e);
  return r !== void 0 ? r : ye(t, e);
}
function pt(n, t, e) {
  for (const r in t)
    r !== "__proto__" &&
      r !== "constructor" &&
      (r in n
        ? typeof n[r] == "string" ||
          n[r] instanceof String ||
          typeof t[r] == "string" ||
          t[r] instanceof String
          ? e && (n[r] = t[r])
          : pt(n[r], t[r], e)
        : (n[r] = t[r]));
  return n;
}
function te(n) {
  return n.replace(/[\-\[\]\/\{\}\(\)\*\+\?\.\\\^\$\|]/g, "\\$&");
}
var xn = {
  "&": "&amp;",
  "<": "&lt;",
  ">": "&gt;",
  '"': "&quot;",
  "'": "&#39;",
  "/": "&#x2F;",
};
function bn(n) {
  return typeof n == "string" ? n.replace(/[&<>"'\/]/g, (t) => xn[t]) : n;
}
class vn {
  constructor(t) {
    ((this.capacity = t),
      (this.regExpMap = new Map()),
      (this.regExpQueue = []));
  }
  getRegExp(t) {
    const e = this.regExpMap.get(t);
    if (e !== void 0) return e;
    const r = new RegExp(t);
    return (
      this.regExpQueue.length === this.capacity &&
        this.regExpMap.delete(this.regExpQueue.shift()),
      this.regExpMap.set(t, r),
      this.regExpQueue.push(t),
      r
    );
  }
}
const yn = [" ", ",", "?", "!", ";"],
  _n = new vn(20);
function wn(n, t, e) {
  ((t = t || ""), (e = e || ""));
  const r = yn.filter((o) => t.indexOf(o) < 0 && e.indexOf(o) < 0);
  if (r.length === 0) return !0;
  const s = _n.getRegExp(
    `(${r.map((o) => (o === "?" ? "\\?" : o)).join("|")})`,
  );
  let i = !s.test(n);
  if (!i) {
    const o = n.indexOf(e);
    o > 0 && !s.test(n.substring(0, o)) && (i = !0);
  }
  return i;
}
function Pe(n, t) {
  let e = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : ".";
  if (!n) return;
  if (n[t]) return n[t];
  const r = t.split(e);
  let s = n;
  for (let i = 0; i < r.length; ) {
    if (!s || typeof s != "object") return;
    let o,
      l = "";
    for (let a = i; a < r.length; ++a)
      if ((a !== i && (l += e), (l += r[a]), (o = s[l]), o !== void 0)) {
        if (
          ["string", "number", "boolean"].indexOf(typeof o) > -1 &&
          a < r.length - 1
        )
          continue;
        i += a - i + 1;
        break;
      }
    s = o;
  }
  return s;
}
function _e(n) {
  return n && n.indexOf("_") > 0 ? n.replace("_", "-") : n;
}
class Ye extends Ce {
  constructor(t) {
    let e =
      arguments.length > 1 && arguments[1] !== void 0
        ? arguments[1]
        : { ns: ["translation"], defaultNS: "translation" };
    (super(),
      (this.data = t || {}),
      (this.options = e),
      this.options.keySeparator === void 0 && (this.options.keySeparator = "."),
      this.options.ignoreJSONStructure === void 0 &&
        (this.options.ignoreJSONStructure = !0));
  }
  addNamespaces(t) {
    this.options.ns.indexOf(t) < 0 && this.options.ns.push(t);
  }
  removeNamespaces(t) {
    const e = this.options.ns.indexOf(t);
    e > -1 && this.options.ns.splice(e, 1);
  }
  getResource(t, e, r) {
    let s = arguments.length > 3 && arguments[3] !== void 0 ? arguments[3] : {};
    const i =
        s.keySeparator !== void 0 ? s.keySeparator : this.options.keySeparator,
      o =
        s.ignoreJSONStructure !== void 0
          ? s.ignoreJSONStructure
          : this.options.ignoreJSONStructure;
    let l;
    t.indexOf(".") > -1
      ? (l = t.split("."))
      : ((l = [t, e]),
        r &&
          (Array.isArray(r)
            ? l.push(...r)
            : typeof r == "string" && i
              ? l.push(...r.split(i))
              : l.push(r)));
    const a = ye(this.data, l);
    return (
      !a &&
        !e &&
        !r &&
        t.indexOf(".") > -1 &&
        ((t = l[0]), (e = l[1]), (r = l.slice(2).join("."))),
      a || !o || typeof r != "string"
        ? a
        : Pe(this.data && this.data[t] && this.data[t][e], r, i)
    );
  }
  addResource(t, e, r, s) {
    let i =
      arguments.length > 4 && arguments[4] !== void 0
        ? arguments[4]
        : { silent: !1 };
    const o =
      i.keySeparator !== void 0 ? i.keySeparator : this.options.keySeparator;
    let l = [t, e];
    (r && (l = l.concat(o ? r.split(o) : r)),
      t.indexOf(".") > -1 && ((l = t.split(".")), (s = e), (e = l[1])),
      this.addNamespaces(e),
      He(this.data, l, s),
      i.silent || this.emit("added", t, e, r, s));
  }
  addResources(t, e, r) {
    let s =
      arguments.length > 3 && arguments[3] !== void 0
        ? arguments[3]
        : { silent: !1 };
    for (const i in r)
      (typeof r[i] == "string" ||
        Object.prototype.toString.apply(r[i]) === "[object Array]") &&
        this.addResource(t, e, i, r[i], { silent: !0 });
    s.silent || this.emit("added", t, e, r);
  }
  addResourceBundle(t, e, r, s, i) {
    let o =
        arguments.length > 5 && arguments[5] !== void 0
          ? arguments[5]
          : { silent: !1, skipCopy: !1 },
      l = [t, e];
    (t.indexOf(".") > -1 && ((l = t.split(".")), (s = r), (r = e), (e = l[1])),
      this.addNamespaces(e));
    let a = ye(this.data, l) || {};
    (o.skipCopy || (r = JSON.parse(JSON.stringify(r))),
      s ? pt(a, r, i) : (a = { ...a, ...r }),
      He(this.data, l, a),
      o.silent || this.emit("added", t, e, r));
  }
  removeResourceBundle(t, e) {
    (this.hasResourceBundle(t, e) && delete this.data[t][e],
      this.removeNamespaces(e),
      this.emit("removed", t, e));
  }
  hasResourceBundle(t, e) {
    return this.getResource(t, e) !== void 0;
  }
  getResourceBundle(t, e) {
    return (
      e || (e = this.options.defaultNS),
      this.options.compatibilityAPI === "v1"
        ? { ...this.getResource(t, e) }
        : this.getResource(t, e)
    );
  }
  getDataByLanguage(t) {
    return this.data[t];
  }
  hasLanguageSomeTranslations(t) {
    const e = this.getDataByLanguage(t);
    return !!((e && Object.keys(e)) || []).find(
      (s) => e[s] && Object.keys(e[s]).length > 0,
    );
  }
  toJSON() {
    return this.data;
  }
}
var gt = {
  processors: {},
  addPostProcessor(n) {
    this.processors[n.name] = n;
  },
  handle(n, t, e, r, s) {
    return (
      n.forEach((i) => {
        this.processors[i] && (t = this.processors[i].process(t, e, r, s));
      }),
      t
    );
  },
};
const Ke = {};
class we extends Ce {
  constructor(t) {
    let e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
    (super(),
      hn(
        [
          "resourceStore",
          "languageUtils",
          "pluralResolver",
          "interpolator",
          "backendConnector",
          "i18nFormat",
          "utils",
        ],
        t,
        this,
      ),
      (this.options = e),
      this.options.keySeparator === void 0 && (this.options.keySeparator = "."),
      (this.logger = Q.create("translator")));
  }
  changeLanguage(t) {
    t && (this.language = t);
  }
  exists(t) {
    let e =
      arguments.length > 1 && arguments[1] !== void 0
        ? arguments[1]
        : { interpolation: {} };
    if (t == null) return !1;
    const r = this.resolve(t, e);
    return r && r.res !== void 0;
  }
  extractFromKey(t, e) {
    let r = e.nsSeparator !== void 0 ? e.nsSeparator : this.options.nsSeparator;
    r === void 0 && (r = ":");
    const s =
      e.keySeparator !== void 0 ? e.keySeparator : this.options.keySeparator;
    let i = e.ns || this.options.defaultNS || [];
    const o = r && t.indexOf(r) > -1,
      l =
        !this.options.userDefinedKeySeparator &&
        !e.keySeparator &&
        !this.options.userDefinedNsSeparator &&
        !e.nsSeparator &&
        !wn(t, r, s);
    if (o && !l) {
      const a = t.match(this.interpolator.nestingRegexp);
      if (a && a.length > 0) return { key: t, namespaces: i };
      const c = t.split(r);
      ((r !== s || (r === s && this.options.ns.indexOf(c[0]) > -1)) &&
        (i = c.shift()),
        (t = c.join(s)));
    }
    return (typeof i == "string" && (i = [i]), { key: t, namespaces: i });
  }
  translate(t, e, r) {
    if (
      (typeof e != "object" &&
        this.options.overloadTranslationOptionHandler &&
        (e = this.options.overloadTranslationOptionHandler(arguments)),
      typeof e == "object" && (e = { ...e }),
      e || (e = {}),
      t == null)
    )
      return "";
    Array.isArray(t) || (t = [String(t)]);
    const s =
        e.returnDetails !== void 0
          ? e.returnDetails
          : this.options.returnDetails,
      i =
        e.keySeparator !== void 0 ? e.keySeparator : this.options.keySeparator,
      { key: o, namespaces: l } = this.extractFromKey(t[t.length - 1], e),
      a = l[l.length - 1],
      c = e.lng || this.language,
      m = e.appendNamespaceToCIMode || this.options.appendNamespaceToCIMode;
    if (c && c.toLowerCase() === "cimode") {
      if (m) {
        const L = e.nsSeparator || this.options.nsSeparator;
        return s
          ? {
              res: `${a}${L}${o}`,
              usedKey: o,
              exactUsedKey: o,
              usedLng: c,
              usedNS: a,
              usedParams: this.getUsedParamsDetails(e),
            }
          : `${a}${L}${o}`;
      }
      return s
        ? {
            res: o,
            usedKey: o,
            exactUsedKey: o,
            usedLng: c,
            usedNS: a,
            usedParams: this.getUsedParamsDetails(e),
          }
        : o;
    }
    const v = this.resolve(t, e);
    let d = v && v.res;
    const w = (v && v.usedKey) || o,
      E = (v && v.exactUsedKey) || o,
      p = Object.prototype.toString.apply(d),
      x = ["[object Number]", "[object Function]", "[object RegExp]"],
      f = e.joinArrays !== void 0 ? e.joinArrays : this.options.joinArrays,
      h = !this.i18nFormat || this.i18nFormat.handleAsObject;
    if (
      h &&
      d &&
      typeof d != "string" &&
      typeof d != "boolean" &&
      typeof d != "number" &&
      x.indexOf(p) < 0 &&
      !(typeof f == "string" && p === "[object Array]")
    ) {
      if (!e.returnObjects && !this.options.returnObjects) {
        this.options.returnedObjectHandler ||
          this.logger.warn(
            "accessing an object - but returnObjects options is not enabled!",
          );
        const L = this.options.returnedObjectHandler
          ? this.options.returnedObjectHandler(w, d, { ...e, ns: l })
          : `key '${o} (${this.language})' returned an object instead of string.`;
        return s
          ? ((v.res = L), (v.usedParams = this.getUsedParamsDetails(e)), v)
          : L;
      }
      if (i) {
        const L = p === "[object Array]",
          W = L ? [] : {},
          u = L ? E : w;
        for (const g in d)
          if (Object.prototype.hasOwnProperty.call(d, g)) {
            const T = `${u}${i}${g}`;
            ((W[g] = this.translate(T, { ...e, joinArrays: !1, ns: l })),
              W[g] === T && (W[g] = d[g]));
          }
        d = W;
      }
    } else if (h && typeof f == "string" && p === "[object Array]")
      ((d = d.join(f)), d && (d = this.extendTranslation(d, t, e, r)));
    else {
      let L = !1,
        W = !1;
      const u = e.count !== void 0 && typeof e.count != "string",
        g = we.hasDefaultValue(e),
        T = u ? this.pluralResolver.getSuffix(c, e.count, e) : "",
        S =
          e.ordinal && u
            ? this.pluralResolver.getSuffix(c, e.count, { ordinal: !1 })
            : "",
        R =
          u &&
          !e.ordinal &&
          e.count === 0 &&
          this.pluralResolver.shouldUseIntlApi(),
        y =
          (R && e[`defaultValue${this.options.pluralSeparator}zero`]) ||
          e[`defaultValue${T}`] ||
          e[`defaultValue${S}`] ||
          e.defaultValue;
      (!this.isValidLookup(d) && g && ((L = !0), (d = y)),
        this.isValidLookup(d) || ((W = !0), (d = o)));
      const _ =
          (e.missingKeyNoValueFallbackToKey ||
            this.options.missingKeyNoValueFallbackToKey) &&
          W
            ? void 0
            : d,
        A = g && y !== d && this.options.updateMissing;
      if (W || L || A) {
        if (
          (this.logger.log(A ? "updateKey" : "missingKey", c, a, o, A ? y : d),
          i)
        ) {
          const I = this.resolve(o, { ...e, keySeparator: !1 });
          I &&
            I.res &&
            this.logger.warn(
              "Seems the loaded translations were in flat JSON format instead of nested. Either set keySeparator: false on init or make sure your translations are published in nested format.",
            );
        }
        let N = [];
        const U = this.languageUtils.getFallbackCodes(
          this.options.fallbackLng,
          e.lng || this.language,
        );
        if (this.options.saveMissingTo === "fallback" && U && U[0])
          for (let I = 0; I < U.length; I++) N.push(U[I]);
        else
          this.options.saveMissingTo === "all"
            ? (N = this.languageUtils.toResolveHierarchy(
                e.lng || this.language,
              ))
            : N.push(e.lng || this.language);
        const $ = (I, O, F) => {
          const M = g && F !== d ? F : _;
          (this.options.missingKeyHandler
            ? this.options.missingKeyHandler(I, a, O, M, A, e)
            : this.backendConnector &&
              this.backendConnector.saveMissing &&
              this.backendConnector.saveMissing(I, a, O, M, A, e),
            this.emit("missingKey", I, a, O, d));
        };
        this.options.saveMissing &&
          (this.options.saveMissingPlurals && u
            ? N.forEach((I) => {
                const O = this.pluralResolver.getSuffixes(I, e);
                (R &&
                  e[`defaultValue${this.options.pluralSeparator}zero`] &&
                  O.indexOf(`${this.options.pluralSeparator}zero`) < 0 &&
                  O.push(`${this.options.pluralSeparator}zero`),
                  O.forEach((F) => {
                    $([I], o + F, e[`defaultValue${F}`] || y);
                  }));
              })
            : $(N, o, y));
      }
      ((d = this.extendTranslation(d, t, e, v, r)),
        W &&
          d === o &&
          this.options.appendNamespaceToMissingKey &&
          (d = `${a}:${o}`),
        (W || L) &&
          this.options.parseMissingKeyHandler &&
          (this.options.compatibilityAPI !== "v1"
            ? (d = this.options.parseMissingKeyHandler(
                this.options.appendNamespaceToMissingKey ? `${a}:${o}` : o,
                L ? d : void 0,
              ))
            : (d = this.options.parseMissingKeyHandler(d))));
    }
    return s
      ? ((v.res = d), (v.usedParams = this.getUsedParamsDetails(e)), v)
      : d;
  }
  extendTranslation(t, e, r, s, i) {
    var o = this;
    if (this.i18nFormat && this.i18nFormat.parse)
      t = this.i18nFormat.parse(
        t,
        { ...this.options.interpolation.defaultVariables, ...r },
        r.lng || this.language || s.usedLng,
        s.usedNS,
        s.usedKey,
        { resolved: s },
      );
    else if (!r.skipInterpolation) {
      r.interpolation &&
        this.interpolator.init({
          ...r,
          interpolation: { ...this.options.interpolation, ...r.interpolation },
        });
      const c =
        typeof t == "string" &&
        (r && r.interpolation && r.interpolation.skipOnVariables !== void 0
          ? r.interpolation.skipOnVariables
          : this.options.interpolation.skipOnVariables);
      let m;
      if (c) {
        const d = t.match(this.interpolator.nestingRegexp);
        m = d && d.length;
      }
      let v = r.replace && typeof r.replace != "string" ? r.replace : r;
      if (
        (this.options.interpolation.defaultVariables &&
          (v = { ...this.options.interpolation.defaultVariables, ...v }),
        (t = this.interpolator.interpolate(t, v, r.lng || this.language, r)),
        c)
      ) {
        const d = t.match(this.interpolator.nestingRegexp),
          w = d && d.length;
        m < w && (r.nest = !1);
      }
      (!r.lng &&
        this.options.compatibilityAPI !== "v1" &&
        s &&
        s.res &&
        (r.lng = s.usedLng),
        r.nest !== !1 &&
          (t = this.interpolator.nest(
            t,
            function () {
              for (
                var d = arguments.length, w = new Array(d), E = 0;
                E < d;
                E++
              )
                w[E] = arguments[E];
              return i && i[0] === w[0] && !r.context
                ? (o.logger.warn(
                    `It seems you are nesting recursively key: ${w[0]} in key: ${e[0]}`,
                  ),
                  null)
                : o.translate(...w, e);
            },
            r,
          )),
        r.interpolation && this.interpolator.reset());
    }
    const l = r.postProcess || this.options.postProcess,
      a = typeof l == "string" ? [l] : l;
    return (
      t != null &&
        a &&
        a.length &&
        r.applyPostProcessor !== !1 &&
        (t = gt.handle(
          a,
          t,
          e,
          this.options && this.options.postProcessPassResolved
            ? {
                i18nResolved: {
                  ...s,
                  usedParams: this.getUsedParamsDetails(r),
                },
                ...r,
              }
            : r,
          this,
        )),
      t
    );
  }
  resolve(t) {
    let e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {},
      r,
      s,
      i,
      o,
      l;
    return (
      typeof t == "string" && (t = [t]),
      t.forEach((a) => {
        if (this.isValidLookup(r)) return;
        const c = this.extractFromKey(a, e),
          m = c.key;
        s = m;
        let v = c.namespaces;
        this.options.fallbackNS && (v = v.concat(this.options.fallbackNS));
        const d = e.count !== void 0 && typeof e.count != "string",
          w =
            d &&
            !e.ordinal &&
            e.count === 0 &&
            this.pluralResolver.shouldUseIntlApi(),
          E =
            e.context !== void 0 &&
            (typeof e.context == "string" || typeof e.context == "number") &&
            e.context !== "",
          p = e.lngs
            ? e.lngs
            : this.languageUtils.toResolveHierarchy(
                e.lng || this.language,
                e.fallbackLng,
              );
        v.forEach((x) => {
          this.isValidLookup(r) ||
            ((l = x),
            !Ke[`${p[0]}-${x}`] &&
              this.utils &&
              this.utils.hasLoadedNamespace &&
              !this.utils.hasLoadedNamespace(l) &&
              ((Ke[`${p[0]}-${x}`] = !0),
              this.logger.warn(
                `key "${s}" for languages "${p.join(", ")}" won't get resolved as namespace "${l}" was not yet loaded`,
                "This means something IS WRONG in your setup. You access the t function before i18next.init / i18next.loadNamespace / i18next.changeLanguage was done. Wait for the callback or Promise to resolve before accessing it!!!",
              )),
            p.forEach((f) => {
              if (this.isValidLookup(r)) return;
              o = f;
              const h = [m];
              if (this.i18nFormat && this.i18nFormat.addLookupKeys)
                this.i18nFormat.addLookupKeys(h, m, f, x, e);
              else {
                let L;
                d && (L = this.pluralResolver.getSuffix(f, e.count, e));
                const W = `${this.options.pluralSeparator}zero`,
                  u = `${this.options.pluralSeparator}ordinal${this.options.pluralSeparator}`;
                if (
                  (d &&
                    (h.push(m + L),
                    e.ordinal &&
                      L.indexOf(u) === 0 &&
                      h.push(m + L.replace(u, this.options.pluralSeparator)),
                    w && h.push(m + W)),
                  E)
                ) {
                  const g = `${m}${this.options.contextSeparator}${e.context}`;
                  (h.push(g),
                    d &&
                      (h.push(g + L),
                      e.ordinal &&
                        L.indexOf(u) === 0 &&
                        h.push(g + L.replace(u, this.options.pluralSeparator)),
                      w && h.push(g + W)));
                }
              }
              let b;
              for (; (b = h.pop()); )
                this.isValidLookup(r) ||
                  ((i = b), (r = this.getResource(f, x, b, e)));
            }));
        });
      }),
      { res: r, usedKey: s, exactUsedKey: i, usedLng: o, usedNS: l }
    );
  }
  isValidLookup(t) {
    return (
      t !== void 0 &&
      !(!this.options.returnNull && t === null) &&
      !(!this.options.returnEmptyString && t === "")
    );
  }
  getResource(t, e, r) {
    let s = arguments.length > 3 && arguments[3] !== void 0 ? arguments[3] : {};
    return this.i18nFormat && this.i18nFormat.getResource
      ? this.i18nFormat.getResource(t, e, r, s)
      : this.resourceStore.getResource(t, e, r, s);
  }
  getUsedParamsDetails() {
    let t = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
    const e = [
        "defaultValue",
        "ordinal",
        "context",
        "replace",
        "lng",
        "lngs",
        "fallbackLng",
        "ns",
        "keySeparator",
        "nsSeparator",
        "returnObjects",
        "returnDetails",
        "joinArrays",
        "postProcess",
        "interpolation",
      ],
      r = t.replace && typeof t.replace != "string";
    let s = r ? t.replace : t;
    if (
      (r && typeof t.count < "u" && (s.count = t.count),
      this.options.interpolation.defaultVariables &&
        (s = { ...this.options.interpolation.defaultVariables, ...s }),
      !r)
    ) {
      s = { ...s };
      for (const i of e) delete s[i];
    }
    return s;
  }
  static hasDefaultValue(t) {
    const e = "defaultValue";
    for (const r in t)
      if (
        Object.prototype.hasOwnProperty.call(t, r) &&
        e === r.substring(0, e.length) &&
        t[r] !== void 0
      )
        return !0;
    return !1;
  }
}
function Re(n) {
  return n.charAt(0).toUpperCase() + n.slice(1);
}
class qe {
  constructor(t) {
    ((this.options = t),
      (this.supportedLngs = this.options.supportedLngs || !1),
      (this.logger = Q.create("languageUtils")));
  }
  getScriptPartFromCode(t) {
    if (((t = _e(t)), !t || t.indexOf("-") < 0)) return null;
    const e = t.split("-");
    return e.length === 2 || (e.pop(), e[e.length - 1].toLowerCase() === "x")
      ? null
      : this.formatLanguageCode(e.join("-"));
  }
  getLanguagePartFromCode(t) {
    if (((t = _e(t)), !t || t.indexOf("-") < 0)) return t;
    const e = t.split("-");
    return this.formatLanguageCode(e[0]);
  }
  formatLanguageCode(t) {
    if (typeof t == "string" && t.indexOf("-") > -1) {
      const e = ["hans", "hant", "latn", "cyrl", "cans", "mong", "arab"];
      let r = t.split("-");
      return (
        this.options.lowerCaseLng
          ? (r = r.map((s) => s.toLowerCase()))
          : r.length === 2
            ? ((r[0] = r[0].toLowerCase()),
              (r[1] = r[1].toUpperCase()),
              e.indexOf(r[1].toLowerCase()) > -1 &&
                (r[1] = Re(r[1].toLowerCase())))
            : r.length === 3 &&
              ((r[0] = r[0].toLowerCase()),
              r[1].length === 2 && (r[1] = r[1].toUpperCase()),
              r[0] !== "sgn" &&
                r[2].length === 2 &&
                (r[2] = r[2].toUpperCase()),
              e.indexOf(r[1].toLowerCase()) > -1 &&
                (r[1] = Re(r[1].toLowerCase())),
              e.indexOf(r[2].toLowerCase()) > -1 &&
                (r[2] = Re(r[2].toLowerCase()))),
        r.join("-")
      );
    }
    return this.options.cleanCode || this.options.lowerCaseLng
      ? t.toLowerCase()
      : t;
  }
  isSupportedCode(t) {
    return (
      (this.options.load === "languageOnly" ||
        this.options.nonExplicitSupportedLngs) &&
        (t = this.getLanguagePartFromCode(t)),
      !this.supportedLngs ||
        !this.supportedLngs.length ||
        this.supportedLngs.indexOf(t) > -1
    );
  }
  getBestMatchFromCodes(t) {
    if (!t) return null;
    let e;
    return (
      t.forEach((r) => {
        if (e) return;
        const s = this.formatLanguageCode(r);
        (!this.options.supportedLngs || this.isSupportedCode(s)) && (e = s);
      }),
      !e &&
        this.options.supportedLngs &&
        t.forEach((r) => {
          if (e) return;
          const s = this.getLanguagePartFromCode(r);
          if (this.isSupportedCode(s)) return (e = s);
          e = this.options.supportedLngs.find((i) => {
            if (i === s) return i;
            if (
              !(i.indexOf("-") < 0 && s.indexOf("-") < 0) &&
              ((i.indexOf("-") > 0 &&
                s.indexOf("-") < 0 &&
                i.substring(0, i.indexOf("-")) === s) ||
                (i.indexOf(s) === 0 && s.length > 1))
            )
              return i;
          });
        }),
      e || (e = this.getFallbackCodes(this.options.fallbackLng)[0]),
      e
    );
  }
  getFallbackCodes(t, e) {
    if (!t) return [];
    if (
      (typeof t == "function" && (t = t(e)),
      typeof t == "string" && (t = [t]),
      Object.prototype.toString.apply(t) === "[object Array]")
    )
      return t;
    if (!e) return t.default || [];
    let r = t[e];
    return (
      r || (r = t[this.getScriptPartFromCode(e)]),
      r || (r = t[this.formatLanguageCode(e)]),
      r || (r = t[this.getLanguagePartFromCode(e)]),
      r || (r = t.default),
      r || []
    );
  }
  toResolveHierarchy(t, e) {
    const r = this.getFallbackCodes(e || this.options.fallbackLng || [], t),
      s = [],
      i = (o) => {
        o &&
          (this.isSupportedCode(o)
            ? s.push(o)
            : this.logger.warn(
                `rejecting language code not found in supportedLngs: ${o}`,
              ));
      };
    return (
      typeof t == "string" && (t.indexOf("-") > -1 || t.indexOf("_") > -1)
        ? (this.options.load !== "languageOnly" &&
            i(this.formatLanguageCode(t)),
          this.options.load !== "languageOnly" &&
            this.options.load !== "currentOnly" &&
            i(this.getScriptPartFromCode(t)),
          this.options.load !== "currentOnly" &&
            i(this.getLanguagePartFromCode(t)))
        : typeof t == "string" && i(this.formatLanguageCode(t)),
      r.forEach((o) => {
        s.indexOf(o) < 0 && i(this.formatLanguageCode(o));
      }),
      s
    );
  }
}
let Sn = [
    {
      lngs: [
        "ach",
        "ak",
        "am",
        "arn",
        "br",
        "fil",
        "gun",
        "ln",
        "mfe",
        "mg",
        "mi",
        "oc",
        "pt",
        "pt-BR",
        "tg",
        "tl",
        "ti",
        "tr",
        "uz",
        "wa",
      ],
      nr: [1, 2],
      fc: 1,
    },
    {
      lngs: [
        "af",
        "an",
        "ast",
        "az",
        "bg",
        "bn",
        "ca",
        "da",
        "de",
        "dev",
        "el",
        "en",
        "eo",
        "es",
        "et",
        "eu",
        "fi",
        "fo",
        "fur",
        "fy",
        "gl",
        "gu",
        "ha",
        "hi",
        "hu",
        "hy",
        "ia",
        "it",
        "kk",
        "kn",
        "ku",
        "lb",
        "mai",
        "ml",
        "mn",
        "mr",
        "nah",
        "nap",
        "nb",
        "ne",
        "nl",
        "nn",
        "no",
        "nso",
        "pa",
        "pap",
        "pms",
        "ps",
        "pt-PT",
        "rm",
        "sco",
        "se",
        "si",
        "so",
        "son",
        "sq",
        "sv",
        "sw",
        "ta",
        "te",
        "tk",
        "ur",
        "yo",
      ],
      nr: [1, 2],
      fc: 2,
    },
    {
      lngs: [
        "ay",
        "bo",
        "cgg",
        "fa",
        "ht",
        "id",
        "ja",
        "jbo",
        "ka",
        "km",
        "ko",
        "ky",
        "lo",
        "ms",
        "sah",
        "su",
        "th",
        "tt",
        "ug",
        "vi",
        "wo",
        "zh",
      ],
      nr: [1],
      fc: 3,
    },
    {
      lngs: ["be", "bs", "cnr", "dz", "hr", "ru", "sr", "uk"],
      nr: [1, 2, 5],
      fc: 4,
    },
    { lngs: ["ar"], nr: [0, 1, 2, 3, 11, 100], fc: 5 },
    { lngs: ["cs", "sk"], nr: [1, 2, 5], fc: 6 },
    { lngs: ["csb", "pl"], nr: [1, 2, 5], fc: 7 },
    { lngs: ["cy"], nr: [1, 2, 3, 8], fc: 8 },
    { lngs: ["fr"], nr: [1, 2], fc: 9 },
    { lngs: ["ga"], nr: [1, 2, 3, 7, 11], fc: 10 },
    { lngs: ["gd"], nr: [1, 2, 3, 20], fc: 11 },
    { lngs: ["is"], nr: [1, 2], fc: 12 },
    { lngs: ["jv"], nr: [0, 1], fc: 13 },
    { lngs: ["kw"], nr: [1, 2, 3, 4], fc: 14 },
    { lngs: ["lt"], nr: [1, 2, 10], fc: 15 },
    { lngs: ["lv"], nr: [1, 2, 0], fc: 16 },
    { lngs: ["mk"], nr: [1, 2], fc: 17 },
    { lngs: ["mnk"], nr: [0, 1, 2], fc: 18 },
    { lngs: ["mt"], nr: [1, 2, 11, 20], fc: 19 },
    { lngs: ["or"], nr: [2, 1], fc: 2 },
    { lngs: ["ro"], nr: [1, 2, 20], fc: 20 },
    { lngs: ["sl"], nr: [5, 1, 2, 3], fc: 21 },
    { lngs: ["he", "iw"], nr: [1, 2, 20, 21], fc: 22 },
  ],
  En = {
    1: function (n) {
      return +(n > 1);
    },
    2: function (n) {
      return +(n != 1);
    },
    3: function (n) {
      return 0;
    },
    4: function (n) {
      return n % 10 == 1 && n % 100 != 11
        ? 0
        : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20)
          ? 1
          : 2;
    },
    5: function (n) {
      return n == 0
        ? 0
        : n == 1
          ? 1
          : n == 2
            ? 2
            : n % 100 >= 3 && n % 100 <= 10
              ? 3
              : n % 100 >= 11
                ? 4
                : 5;
    },
    6: function (n) {
      return n == 1 ? 0 : n >= 2 && n <= 4 ? 1 : 2;
    },
    7: function (n) {
      return n == 1
        ? 0
        : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20)
          ? 1
          : 2;
    },
    8: function (n) {
      return n == 1 ? 0 : n == 2 ? 1 : n != 8 && n != 11 ? 2 : 3;
    },
    9: function (n) {
      return +(n >= 2);
    },
    10: function (n) {
      return n == 1 ? 0 : n == 2 ? 1 : n < 7 ? 2 : n < 11 ? 3 : 4;
    },
    11: function (n) {
      return n == 1 || n == 11
        ? 0
        : n == 2 || n == 12
          ? 1
          : n > 2 && n < 20
            ? 2
            : 3;
    },
    12: function (n) {
      return +(n % 10 != 1 || n % 100 == 11);
    },
    13: function (n) {
      return +(n !== 0);
    },
    14: function (n) {
      return n == 1 ? 0 : n == 2 ? 1 : n == 3 ? 2 : 3;
    },
    15: function (n) {
      return n % 10 == 1 && n % 100 != 11
        ? 0
        : n % 10 >= 2 && (n % 100 < 10 || n % 100 >= 20)
          ? 1
          : 2;
    },
    16: function (n) {
      return n % 10 == 1 && n % 100 != 11 ? 0 : n !== 0 ? 1 : 2;
    },
    17: function (n) {
      return n == 1 || (n % 10 == 1 && n % 100 != 11) ? 0 : 1;
    },
    18: function (n) {
      return n == 0 ? 0 : n == 1 ? 1 : 2;
    },
    19: function (n) {
      return n == 1
        ? 0
        : n == 0 || (n % 100 > 1 && n % 100 < 11)
          ? 1
          : n % 100 > 10 && n % 100 < 20
            ? 2
            : 3;
    },
    20: function (n) {
      return n == 1 ? 0 : n == 0 || (n % 100 > 0 && n % 100 < 20) ? 1 : 2;
    },
    21: function (n) {
      return n % 100 == 1
        ? 1
        : n % 100 == 2
          ? 2
          : n % 100 == 3 || n % 100 == 4
            ? 3
            : 0;
    },
    22: function (n) {
      return n == 1 ? 0 : n == 2 ? 1 : (n < 0 || n > 10) && n % 10 == 0 ? 2 : 3;
    },
  };
const On = ["v1", "v2", "v3"],
  Tn = ["v4"],
  Qe = { zero: 0, one: 1, two: 2, few: 3, many: 4, other: 5 };
function Ln() {
  const n = {};
  return (
    Sn.forEach((t) => {
      t.lngs.forEach((e) => {
        n[e] = { numbers: t.nr, plurals: En[t.fc] };
      });
    }),
    n
  );
}
class Cn {
  constructor(t) {
    let e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
    ((this.languageUtils = t),
      (this.options = e),
      (this.logger = Q.create("pluralResolver")),
      (!this.options.compatibilityJSON ||
        Tn.includes(this.options.compatibilityJSON)) &&
        (typeof Intl > "u" || !Intl.PluralRules) &&
        ((this.options.compatibilityJSON = "v3"),
        this.logger.error(
          "Your environment seems not to be Intl API compatible, use an Intl.PluralRules polyfill. Will fallback to the compatibilityJSON v3 format handling.",
        )),
      (this.rules = Ln()));
  }
  addRule(t, e) {
    this.rules[t] = e;
  }
  getRule(t) {
    let e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
    if (this.shouldUseIntlApi())
      try {
        return new Intl.PluralRules(_e(t === "dev" ? "en" : t), {
          type: e.ordinal ? "ordinal" : "cardinal",
        });
      } catch {
        return;
      }
    return (
      this.rules[t] || this.rules[this.languageUtils.getLanguagePartFromCode(t)]
    );
  }
  needsPlural(t) {
    let e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
    const r = this.getRule(t, e);
    return this.shouldUseIntlApi()
      ? r && r.resolvedOptions().pluralCategories.length > 1
      : r && r.numbers.length > 1;
  }
  getPluralFormsOfKey(t, e) {
    let r = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : {};
    return this.getSuffixes(t, r).map((s) => `${e}${s}`);
  }
  getSuffixes(t) {
    let e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
    const r = this.getRule(t, e);
    return r
      ? this.shouldUseIntlApi()
        ? r
            .resolvedOptions()
            .pluralCategories.sort((s, i) => Qe[s] - Qe[i])
            .map(
              (s) =>
                `${this.options.prepend}${e.ordinal ? `ordinal${this.options.prepend}` : ""}${s}`,
            )
        : r.numbers.map((s) => this.getSuffix(t, s, e))
      : [];
  }
  getSuffix(t, e) {
    let r = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : {};
    const s = this.getRule(t, r);
    return s
      ? this.shouldUseIntlApi()
        ? `${this.options.prepend}${r.ordinal ? `ordinal${this.options.prepend}` : ""}${s.select(e)}`
        : this.getSuffixRetroCompatible(s, e)
      : (this.logger.warn(`no plural rule found for: ${t}`), "");
  }
  getSuffixRetroCompatible(t, e) {
    const r = t.noAbs ? t.plurals(e) : t.plurals(Math.abs(e));
    let s = t.numbers[r];
    this.options.simplifyPluralSuffix &&
      t.numbers.length === 2 &&
      t.numbers[0] === 1 &&
      (s === 2 ? (s = "plural") : s === 1 && (s = ""));
    const i = () =>
      this.options.prepend && s.toString()
        ? this.options.prepend + s.toString()
        : s.toString();
    return this.options.compatibilityJSON === "v1"
      ? s === 1
        ? ""
        : typeof s == "number"
          ? `_plural_${s.toString()}`
          : i()
      : this.options.compatibilityJSON === "v2" ||
          (this.options.simplifyPluralSuffix &&
            t.numbers.length === 2 &&
            t.numbers[0] === 1)
        ? i()
        : this.options.prepend && r.toString()
          ? this.options.prepend + r.toString()
          : r.toString();
  }
  shouldUseIntlApi() {
    return !On.includes(this.options.compatibilityJSON);
  }
}
function Ge(n, t, e) {
  let r = arguments.length > 3 && arguments[3] !== void 0 ? arguments[3] : ".",
    s = arguments.length > 4 && arguments[4] !== void 0 ? arguments[4] : !0,
    i = mn(n, t, e);
  return (
    !i &&
      s &&
      typeof e == "string" &&
      ((i = Pe(n, e, r)), i === void 0 && (i = Pe(t, e, r))),
    i
  );
}
class kn {
  constructor() {
    let t = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
    ((this.logger = Q.create("interpolator")),
      (this.options = t),
      (this.format = (t.interpolation && t.interpolation.format) || ((e) => e)),
      this.init(t));
  }
  init() {
    let t = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
    t.interpolation || (t.interpolation = { escapeValue: !0 });
    const e = t.interpolation;
    ((this.escape = e.escape !== void 0 ? e.escape : bn),
      (this.escapeValue = e.escapeValue !== void 0 ? e.escapeValue : !0),
      (this.useRawValueToEscape =
        e.useRawValueToEscape !== void 0 ? e.useRawValueToEscape : !1),
      (this.prefix = e.prefix ? te(e.prefix) : e.prefixEscaped || "{{"),
      (this.suffix = e.suffix ? te(e.suffix) : e.suffixEscaped || "}}"),
      (this.formatSeparator = e.formatSeparator
        ? e.formatSeparator
        : e.formatSeparator || ","),
      (this.unescapePrefix = e.unescapeSuffix ? "" : e.unescapePrefix || "-"),
      (this.unescapeSuffix = this.unescapePrefix ? "" : e.unescapeSuffix || ""),
      (this.nestingPrefix = e.nestingPrefix
        ? te(e.nestingPrefix)
        : e.nestingPrefixEscaped || te("$t(")),
      (this.nestingSuffix = e.nestingSuffix
        ? te(e.nestingSuffix)
        : e.nestingSuffixEscaped || te(")")),
      (this.nestingOptionsSeparator = e.nestingOptionsSeparator
        ? e.nestingOptionsSeparator
        : e.nestingOptionsSeparator || ","),
      (this.maxReplaces = e.maxReplaces ? e.maxReplaces : 1e3),
      (this.alwaysFormat = e.alwaysFormat !== void 0 ? e.alwaysFormat : !1),
      this.resetRegExp());
  }
  reset() {
    this.options && this.init(this.options);
  }
  resetRegExp() {
    const t = (e, r) =>
      e && e.source === r ? ((e.lastIndex = 0), e) : new RegExp(r, "g");
    ((this.regexp = t(this.regexp, `${this.prefix}(.+?)${this.suffix}`)),
      (this.regexpUnescape = t(
        this.regexpUnescape,
        `${this.prefix}${this.unescapePrefix}(.+?)${this.unescapeSuffix}${this.suffix}`,
      )),
      (this.nestingRegexp = t(
        this.nestingRegexp,
        `${this.nestingPrefix}(.+?)${this.nestingSuffix}`,
      )));
  }
  interpolate(t, e, r, s) {
    let i, o, l;
    const a =
      (this.options &&
        this.options.interpolation &&
        this.options.interpolation.defaultVariables) ||
      {};
    function c(E) {
      return E.replace(/\$/g, "$$$$");
    }
    const m = (E) => {
      if (E.indexOf(this.formatSeparator) < 0) {
        const h = Ge(
          e,
          a,
          E,
          this.options.keySeparator,
          this.options.ignoreJSONStructure,
        );
        return this.alwaysFormat
          ? this.format(h, void 0, r, { ...s, ...e, interpolationkey: E })
          : h;
      }
      const p = E.split(this.formatSeparator),
        x = p.shift().trim(),
        f = p.join(this.formatSeparator).trim();
      return this.format(
        Ge(
          e,
          a,
          x,
          this.options.keySeparator,
          this.options.ignoreJSONStructure,
        ),
        f,
        r,
        { ...s, ...e, interpolationkey: x },
      );
    };
    this.resetRegExp();
    const v =
        (s && s.missingInterpolationHandler) ||
        this.options.missingInterpolationHandler,
      d =
        s && s.interpolation && s.interpolation.skipOnVariables !== void 0
          ? s.interpolation.skipOnVariables
          : this.options.interpolation.skipOnVariables;
    return (
      [
        { regex: this.regexpUnescape, safeValue: (E) => c(E) },
        {
          regex: this.regexp,
          safeValue: (E) => (this.escapeValue ? c(this.escape(E)) : c(E)),
        },
      ].forEach((E) => {
        for (l = 0; (i = E.regex.exec(t)); ) {
          const p = i[1].trim();
          if (((o = m(p)), o === void 0))
            if (typeof v == "function") {
              const f = v(t, i, s);
              o = typeof f == "string" ? f : "";
            } else if (s && Object.prototype.hasOwnProperty.call(s, p)) o = "";
            else if (d) {
              o = i[0];
              continue;
            } else
              (this.logger.warn(
                `missed to pass in variable ${p} for interpolating ${t}`,
              ),
                (o = ""));
          else typeof o != "string" && !this.useRawValueToEscape && (o = ze(o));
          const x = E.safeValue(o);
          if (
            ((t = t.replace(i[0], x)),
            d
              ? ((E.regex.lastIndex += o.length),
                (E.regex.lastIndex -= i[0].length))
              : (E.regex.lastIndex = 0),
            l++,
            l >= this.maxReplaces)
          )
            break;
        }
      }),
      t
    );
  }
  nest(t, e) {
    let r = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : {},
      s,
      i,
      o;
    function l(a, c) {
      const m = this.nestingOptionsSeparator;
      if (a.indexOf(m) < 0) return a;
      const v = a.split(new RegExp(`${m}[ ]*{`));
      let d = `{${v[1]}`;
      ((a = v[0]), (d = this.interpolate(d, o)));
      const w = d.match(/'/g),
        E = d.match(/"/g);
      ((w && w.length % 2 === 0 && !E) || E.length % 2 !== 0) &&
        (d = d.replace(/'/g, '"'));
      try {
        ((o = JSON.parse(d)), c && (o = { ...c, ...o }));
      } catch (p) {
        return (
          this.logger.warn(
            `failed parsing options string in nesting for key ${a}`,
            p,
          ),
          `${a}${m}${d}`
        );
      }
      return (delete o.defaultValue, a);
    }
    for (; (s = this.nestingRegexp.exec(t)); ) {
      let a = [];
      ((o = { ...r }),
        (o = o.replace && typeof o.replace != "string" ? o.replace : o),
        (o.applyPostProcessor = !1),
        delete o.defaultValue);
      let c = !1;
      if (s[0].indexOf(this.formatSeparator) !== -1 && !/{.*}/.test(s[1])) {
        const m = s[1].split(this.formatSeparator).map((v) => v.trim());
        ((s[1] = m.shift()), (a = m), (c = !0));
      }
      if (
        ((i = e(l.call(this, s[1].trim(), o), o)),
        i && s[0] === t && typeof i != "string")
      )
        return i;
      (typeof i != "string" && (i = ze(i)),
        i ||
          (this.logger.warn(`missed to resolve ${s[1]} for nesting ${t}`),
          (i = "")),
        c &&
          (i = a.reduce(
            (m, v) =>
              this.format(m, v, r.lng, { ...r, interpolationkey: s[1].trim() }),
            i.trim(),
          )),
        (t = t.replace(s[0], i)),
        (this.regexp.lastIndex = 0));
    }
    return t;
  }
}
function Rn(n) {
  let t = n.toLowerCase().trim();
  const e = {};
  if (n.indexOf("(") > -1) {
    const r = n.split("(");
    t = r[0].toLowerCase().trim();
    const s = r[1].substring(0, r[1].length - 1);
    t === "currency" && s.indexOf(":") < 0
      ? e.currency || (e.currency = s.trim())
      : t === "relativetime" && s.indexOf(":") < 0
        ? e.range || (e.range = s.trim())
        : s.split(";").forEach((o) => {
            if (!o) return;
            const [l, ...a] = o.split(":"),
              c = a
                .join(":")
                .trim()
                .replace(/^'+|'+$/g, "");
            (e[l.trim()] || (e[l.trim()] = c),
              c === "false" && (e[l.trim()] = !1),
              c === "true" && (e[l.trim()] = !0),
              isNaN(c) || (e[l.trim()] = parseInt(c, 10)));
          });
  }
  return { formatName: t, formatOptions: e };
}
function re(n) {
  const t = {};
  return function (r, s, i) {
    const o = s + JSON.stringify(i);
    let l = t[o];
    return (l || ((l = n(_e(s), i)), (t[o] = l)), l(r));
  };
}
class In {
  constructor() {
    let t = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
    ((this.logger = Q.create("formatter")),
      (this.options = t),
      (this.formats = {
        number: re((e, r) => {
          const s = new Intl.NumberFormat(e, { ...r });
          return (i) => s.format(i);
        }),
        currency: re((e, r) => {
          const s = new Intl.NumberFormat(e, { ...r, style: "currency" });
          return (i) => s.format(i);
        }),
        datetime: re((e, r) => {
          const s = new Intl.DateTimeFormat(e, { ...r });
          return (i) => s.format(i);
        }),
        relativetime: re((e, r) => {
          const s = new Intl.RelativeTimeFormat(e, { ...r });
          return (i) => s.format(i, r.range || "day");
        }),
        list: re((e, r) => {
          const s = new Intl.ListFormat(e, { ...r });
          return (i) => s.format(i);
        }),
      }),
      this.init(t));
  }
  init(t) {
    const r = (
      arguments.length > 1 && arguments[1] !== void 0
        ? arguments[1]
        : { interpolation: {} }
    ).interpolation;
    this.formatSeparator = r.formatSeparator
      ? r.formatSeparator
      : r.formatSeparator || ",";
  }
  add(t, e) {
    this.formats[t.toLowerCase().trim()] = e;
  }
  addCached(t, e) {
    this.formats[t.toLowerCase().trim()] = re(e);
  }
  format(t, e, r) {
    let s = arguments.length > 3 && arguments[3] !== void 0 ? arguments[3] : {};
    return e.split(this.formatSeparator).reduce((l, a) => {
      const { formatName: c, formatOptions: m } = Rn(a);
      if (this.formats[c]) {
        let v = l;
        try {
          const d =
              (s && s.formatParams && s.formatParams[s.interpolationkey]) || {},
            w = d.locale || d.lng || s.locale || s.lng || r;
          v = this.formats[c](l, w, { ...m, ...s, ...d });
        } catch (d) {
          this.logger.warn(d);
        }
        return v;
      } else this.logger.warn(`there was no format function for ${c}`);
      return l;
    }, t);
  }
}
function Pn(n, t) {
  n.pending[t] !== void 0 && (delete n.pending[t], n.pendingCount--);
}
class An extends Ce {
  constructor(t, e, r) {
    let s = arguments.length > 3 && arguments[3] !== void 0 ? arguments[3] : {};
    (super(),
      (this.backend = t),
      (this.store = e),
      (this.services = r),
      (this.languageUtils = r.languageUtils),
      (this.options = s),
      (this.logger = Q.create("backendConnector")),
      (this.waitingReads = []),
      (this.maxParallelReads = s.maxParallelReads || 10),
      (this.readingCalls = 0),
      (this.maxRetries = s.maxRetries >= 0 ? s.maxRetries : 5),
      (this.retryTimeout = s.retryTimeout >= 1 ? s.retryTimeout : 350),
      (this.state = {}),
      (this.queue = []),
      this.backend && this.backend.init && this.backend.init(r, s.backend, s));
  }
  queueLoad(t, e, r, s) {
    const i = {},
      o = {},
      l = {},
      a = {};
    return (
      t.forEach((c) => {
        let m = !0;
        (e.forEach((v) => {
          const d = `${c}|${v}`;
          !r.reload && this.store.hasResourceBundle(c, v)
            ? (this.state[d] = 2)
            : this.state[d] < 0 ||
              (this.state[d] === 1
                ? o[d] === void 0 && (o[d] = !0)
                : ((this.state[d] = 1),
                  (m = !1),
                  o[d] === void 0 && (o[d] = !0),
                  i[d] === void 0 && (i[d] = !0),
                  a[v] === void 0 && (a[v] = !0)));
        }),
          m || (l[c] = !0));
      }),
      (Object.keys(i).length || Object.keys(o).length) &&
        this.queue.push({
          pending: o,
          pendingCount: Object.keys(o).length,
          loaded: {},
          errors: [],
          callback: s,
        }),
      {
        toLoad: Object.keys(i),
        pending: Object.keys(o),
        toLoadLanguages: Object.keys(l),
        toLoadNamespaces: Object.keys(a),
      }
    );
  }
  loaded(t, e, r) {
    const s = t.split("|"),
      i = s[0],
      o = s[1];
    (e && this.emit("failedLoading", i, o, e),
      r &&
        this.store.addResourceBundle(i, o, r, void 0, void 0, { skipCopy: !0 }),
      (this.state[t] = e ? -1 : 2));
    const l = {};
    (this.queue.forEach((a) => {
      (gn(a.loaded, [i], o),
        Pn(a, t),
        e && a.errors.push(e),
        a.pendingCount === 0 &&
          !a.done &&
          (Object.keys(a.loaded).forEach((c) => {
            l[c] || (l[c] = {});
            const m = a.loaded[c];
            m.length &&
              m.forEach((v) => {
                l[c][v] === void 0 && (l[c][v] = !0);
              });
          }),
          (a.done = !0),
          a.errors.length ? a.callback(a.errors) : a.callback()));
    }),
      this.emit("loaded", l),
      (this.queue = this.queue.filter((a) => !a.done)));
  }
  read(t, e, r) {
    let s = arguments.length > 3 && arguments[3] !== void 0 ? arguments[3] : 0,
      i =
        arguments.length > 4 && arguments[4] !== void 0
          ? arguments[4]
          : this.retryTimeout,
      o = arguments.length > 5 ? arguments[5] : void 0;
    if (!t.length) return o(null, {});
    if (this.readingCalls >= this.maxParallelReads) {
      this.waitingReads.push({
        lng: t,
        ns: e,
        fcName: r,
        tried: s,
        wait: i,
        callback: o,
      });
      return;
    }
    this.readingCalls++;
    const l = (c, m) => {
        if ((this.readingCalls--, this.waitingReads.length > 0)) {
          const v = this.waitingReads.shift();
          this.read(v.lng, v.ns, v.fcName, v.tried, v.wait, v.callback);
        }
        if (c && m && s < this.maxRetries) {
          setTimeout(() => {
            this.read.call(this, t, e, r, s + 1, i * 2, o);
          }, i);
          return;
        }
        o(c, m);
      },
      a = this.backend[r].bind(this.backend);
    if (a.length === 2) {
      try {
        const c = a(t, e);
        c && typeof c.then == "function"
          ? c.then((m) => l(null, m)).catch(l)
          : l(null, c);
      } catch (c) {
        l(c);
      }
      return;
    }
    return a(t, e, l);
  }
  prepareLoading(t, e) {
    let r = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : {},
      s = arguments.length > 3 ? arguments[3] : void 0;
    if (!this.backend)
      return (
        this.logger.warn(
          "No backend was added via i18next.use. Will not load resources.",
        ),
        s && s()
      );
    (typeof t == "string" && (t = this.languageUtils.toResolveHierarchy(t)),
      typeof e == "string" && (e = [e]));
    const i = this.queueLoad(t, e, r, s);
    if (!i.toLoad.length) return (i.pending.length || s(), null);
    i.toLoad.forEach((o) => {
      this.loadOne(o);
    });
  }
  load(t, e, r) {
    this.prepareLoading(t, e, {}, r);
  }
  reload(t, e, r) {
    this.prepareLoading(t, e, { reload: !0 }, r);
  }
  loadOne(t) {
    let e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : "";
    const r = t.split("|"),
      s = r[0],
      i = r[1];
    this.read(s, i, "read", void 0, void 0, (o, l) => {
      (o &&
        this.logger.warn(
          `${e}loading namespace ${i} for language ${s} failed`,
          o,
        ),
        !o &&
          l &&
          this.logger.log(`${e}loaded namespace ${i} for language ${s}`, l),
        this.loaded(t, o, l));
    });
  }
  saveMissing(t, e, r, s, i) {
    let o = arguments.length > 5 && arguments[5] !== void 0 ? arguments[5] : {},
      l =
        arguments.length > 6 && arguments[6] !== void 0
          ? arguments[6]
          : () => {};
    if (
      this.services.utils &&
      this.services.utils.hasLoadedNamespace &&
      !this.services.utils.hasLoadedNamespace(e)
    ) {
      this.logger.warn(
        `did not save key "${r}" as the namespace "${e}" was not yet loaded`,
        "This means something IS WRONG in your setup. You access the t function before i18next.init / i18next.loadNamespace / i18next.changeLanguage was done. Wait for the callback or Promise to resolve before accessing it!!!",
      );
      return;
    }
    if (!(r == null || r === "")) {
      if (this.backend && this.backend.create) {
        const a = { ...o, isUpdate: i },
          c = this.backend.create.bind(this.backend);
        if (c.length < 6)
          try {
            let m;
            (c.length === 5 ? (m = c(t, e, r, s, a)) : (m = c(t, e, r, s)),
              m && typeof m.then == "function"
                ? m.then((v) => l(null, v)).catch(l)
                : l(null, m));
          } catch (m) {
            l(m);
          }
        else c(t, e, r, s, l, a);
      }
      !t || !t[0] || this.store.addResource(t[0], e, r, s);
    }
  }
}
function Je() {
  return {
    debug: !1,
    initImmediate: !0,
    ns: ["translation"],
    defaultNS: ["translation"],
    fallbackLng: ["dev"],
    fallbackNS: !1,
    supportedLngs: !1,
    nonExplicitSupportedLngs: !1,
    load: "all",
    preload: !1,
    simplifyPluralSuffix: !0,
    keySeparator: ".",
    nsSeparator: ":",
    pluralSeparator: "_",
    contextSeparator: "_",
    partialBundledLanguages: !1,
    saveMissing: !1,
    updateMissing: !1,
    saveMissingTo: "fallback",
    saveMissingPlurals: !0,
    missingKeyHandler: !1,
    missingInterpolationHandler: !1,
    postProcess: !1,
    postProcessPassResolved: !1,
    returnNull: !1,
    returnEmptyString: !0,
    returnObjects: !1,
    joinArrays: !1,
    returnedObjectHandler: !1,
    parseMissingKeyHandler: !1,
    appendNamespaceToMissingKey: !1,
    appendNamespaceToCIMode: !1,
    overloadTranslationOptionHandler: function (t) {
      let e = {};
      if (
        (typeof t[1] == "object" && (e = t[1]),
        typeof t[1] == "string" && (e.defaultValue = t[1]),
        typeof t[2] == "string" && (e.tDescription = t[2]),
        typeof t[2] == "object" || typeof t[3] == "object")
      ) {
        const r = t[3] || t[2];
        Object.keys(r).forEach((s) => {
          e[s] = r[s];
        });
      }
      return e;
    },
    interpolation: {
      escapeValue: !0,
      format: (n) => n,
      prefix: "{{",
      suffix: "}}",
      formatSeparator: ",",
      unescapePrefix: "-",
      nestingPrefix: "$t(",
      nestingSuffix: ")",
      nestingOptionsSeparator: ",",
      maxReplaces: 1e3,
      skipOnVariables: !0,
    },
  };
}
function Xe(n) {
  return (
    typeof n.ns == "string" && (n.ns = [n.ns]),
    typeof n.fallbackLng == "string" && (n.fallbackLng = [n.fallbackLng]),
    typeof n.fallbackNS == "string" && (n.fallbackNS = [n.fallbackNS]),
    n.supportedLngs &&
      n.supportedLngs.indexOf("cimode") < 0 &&
      (n.supportedLngs = n.supportedLngs.concat(["cimode"])),
    n
  );
}
function xe() {}
function Fn(n) {
  Object.getOwnPropertyNames(Object.getPrototypeOf(n)).forEach((e) => {
    typeof n[e] == "function" && (n[e] = n[e].bind(n));
  });
}
class fe extends Ce {
  constructor() {
    let t = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {},
      e = arguments.length > 1 ? arguments[1] : void 0;
    if (
      (super(),
      (this.options = Xe(t)),
      (this.services = {}),
      (this.logger = Q),
      (this.modules = { external: [] }),
      Fn(this),
      e && !this.isInitialized && !t.isClone)
    ) {
      if (!this.options.initImmediate) return (this.init(t, e), this);
      setTimeout(() => {
        this.init(t, e);
      }, 0);
    }
  }
  init() {
    var t = this;
    let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {},
      r = arguments.length > 1 ? arguments[1] : void 0;
    ((this.isInitializing = !0),
      typeof e == "function" && ((r = e), (e = {})),
      !e.defaultNS &&
        e.defaultNS !== !1 &&
        e.ns &&
        (typeof e.ns == "string"
          ? (e.defaultNS = e.ns)
          : e.ns.indexOf("translation") < 0 && (e.defaultNS = e.ns[0])));
    const s = Je();
    ((this.options = { ...s, ...this.options, ...Xe(e) }),
      this.options.compatibilityAPI !== "v1" &&
        (this.options.interpolation = {
          ...s.interpolation,
          ...this.options.interpolation,
        }),
      e.keySeparator !== void 0 &&
        (this.options.userDefinedKeySeparator = e.keySeparator),
      e.nsSeparator !== void 0 &&
        (this.options.userDefinedNsSeparator = e.nsSeparator));
    function i(m) {
      return m ? (typeof m == "function" ? new m() : m) : null;
    }
    if (!this.options.isClone) {
      this.modules.logger
        ? Q.init(i(this.modules.logger), this.options)
        : Q.init(null, this.options);
      let m;
      this.modules.formatter
        ? (m = this.modules.formatter)
        : typeof Intl < "u" && (m = In);
      const v = new qe(this.options);
      this.store = new Ye(this.options.resources, this.options);
      const d = this.services;
      ((d.logger = Q),
        (d.resourceStore = this.store),
        (d.languageUtils = v),
        (d.pluralResolver = new Cn(v, {
          prepend: this.options.pluralSeparator,
          compatibilityJSON: this.options.compatibilityJSON,
          simplifyPluralSuffix: this.options.simplifyPluralSuffix,
        })),
        m &&
          (!this.options.interpolation.format ||
            this.options.interpolation.format === s.interpolation.format) &&
          ((d.formatter = i(m)),
          d.formatter.init(d, this.options),
          (this.options.interpolation.format = d.formatter.format.bind(
            d.formatter,
          ))),
        (d.interpolator = new kn(this.options)),
        (d.utils = { hasLoadedNamespace: this.hasLoadedNamespace.bind(this) }),
        (d.backendConnector = new An(
          i(this.modules.backend),
          d.resourceStore,
          d,
          this.options,
        )),
        d.backendConnector.on("*", function (w) {
          for (
            var E = arguments.length, p = new Array(E > 1 ? E - 1 : 0), x = 1;
            x < E;
            x++
          )
            p[x - 1] = arguments[x];
          t.emit(w, ...p);
        }),
        this.modules.languageDetector &&
          ((d.languageDetector = i(this.modules.languageDetector)),
          d.languageDetector.init &&
            d.languageDetector.init(d, this.options.detection, this.options)),
        this.modules.i18nFormat &&
          ((d.i18nFormat = i(this.modules.i18nFormat)),
          d.i18nFormat.init && d.i18nFormat.init(this)),
        (this.translator = new we(this.services, this.options)),
        this.translator.on("*", function (w) {
          for (
            var E = arguments.length, p = new Array(E > 1 ? E - 1 : 0), x = 1;
            x < E;
            x++
          )
            p[x - 1] = arguments[x];
          t.emit(w, ...p);
        }),
        this.modules.external.forEach((w) => {
          w.init && w.init(this);
        }));
    }
    if (
      ((this.format = this.options.interpolation.format),
      r || (r = xe),
      this.options.fallbackLng &&
        !this.services.languageDetector &&
        !this.options.lng)
    ) {
      const m = this.services.languageUtils.getFallbackCodes(
        this.options.fallbackLng,
      );
      m.length > 0 && m[0] !== "dev" && (this.options.lng = m[0]);
    }
    (!this.services.languageDetector &&
      !this.options.lng &&
      this.logger.warn(
        "init: no languageDetector is used and no lng is defined",
      ),
      [
        "getResource",
        "hasResourceBundle",
        "getResourceBundle",
        "getDataByLanguage",
      ].forEach((m) => {
        this[m] = function () {
          return t.store[m](...arguments);
        };
      }),
      [
        "addResource",
        "addResources",
        "addResourceBundle",
        "removeResourceBundle",
      ].forEach((m) => {
        this[m] = function () {
          return (t.store[m](...arguments), t);
        };
      }));
    const a = ae(),
      c = () => {
        const m = (v, d) => {
          ((this.isInitializing = !1),
            this.isInitialized &&
              !this.initializedStoreOnce &&
              this.logger.warn(
                "init: i18next is already initialized. You should call init just once!",
              ),
            (this.isInitialized = !0),
            this.options.isClone ||
              this.logger.log("initialized", this.options),
            this.emit("initialized", this.options),
            a.resolve(d),
            r(v, d));
        };
        if (
          this.languages &&
          this.options.compatibilityAPI !== "v1" &&
          !this.isInitialized
        )
          return m(null, this.t.bind(this));
        this.changeLanguage(this.options.lng, m);
      };
    return (
      this.options.resources || !this.options.initImmediate
        ? c()
        : setTimeout(c, 0),
      a
    );
  }
  loadResources(t) {
    let r = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : xe;
    const s = typeof t == "string" ? t : this.language;
    if (
      (typeof t == "function" && (r = t),
      !this.options.resources || this.options.partialBundledLanguages)
    ) {
      if (
        s &&
        s.toLowerCase() === "cimode" &&
        (!this.options.preload || this.options.preload.length === 0)
      )
        return r();
      const i = [],
        o = (l) => {
          if (!l || l === "cimode") return;
          this.services.languageUtils.toResolveHierarchy(l).forEach((c) => {
            c !== "cimode" && i.indexOf(c) < 0 && i.push(c);
          });
        };
      (s
        ? o(s)
        : this.services.languageUtils
            .getFallbackCodes(this.options.fallbackLng)
            .forEach((a) => o(a)),
        this.options.preload && this.options.preload.forEach((l) => o(l)),
        this.services.backendConnector.load(i, this.options.ns, (l) => {
          (!l &&
            !this.resolvedLanguage &&
            this.language &&
            this.setResolvedLanguage(this.language),
            r(l));
        }));
    } else r(null);
  }
  reloadResources(t, e, r) {
    const s = ae();
    return (
      t || (t = this.languages),
      e || (e = this.options.ns),
      r || (r = xe),
      this.services.backendConnector.reload(t, e, (i) => {
        (s.resolve(), r(i));
      }),
      s
    );
  }
  use(t) {
    if (!t)
      throw new Error(
        "You are passing an undefined module! Please check the object you are passing to i18next.use()",
      );
    if (!t.type)
      throw new Error(
        "You are passing a wrong module! Please check the object you are passing to i18next.use()",
      );
    return (
      t.type === "backend" && (this.modules.backend = t),
      (t.type === "logger" || (t.log && t.warn && t.error)) &&
        (this.modules.logger = t),
      t.type === "languageDetector" && (this.modules.languageDetector = t),
      t.type === "i18nFormat" && (this.modules.i18nFormat = t),
      t.type === "postProcessor" && gt.addPostProcessor(t),
      t.type === "formatter" && (this.modules.formatter = t),
      t.type === "3rdParty" && this.modules.external.push(t),
      this
    );
  }
  setResolvedLanguage(t) {
    if (!(!t || !this.languages) && !(["cimode", "dev"].indexOf(t) > -1))
      for (let e = 0; e < this.languages.length; e++) {
        const r = this.languages[e];
        if (
          !(["cimode", "dev"].indexOf(r) > -1) &&
          this.store.hasLanguageSomeTranslations(r)
        ) {
          this.resolvedLanguage = r;
          break;
        }
      }
  }
  changeLanguage(t, e) {
    var r = this;
    this.isLanguageChangingTo = t;
    const s = ae();
    this.emit("languageChanging", t);
    const i = (a) => {
        ((this.language = a),
          (this.languages = this.services.languageUtils.toResolveHierarchy(a)),
          (this.resolvedLanguage = void 0),
          this.setResolvedLanguage(a));
      },
      o = (a, c) => {
        (c
          ? (i(c),
            this.translator.changeLanguage(c),
            (this.isLanguageChangingTo = void 0),
            this.emit("languageChanged", c),
            this.logger.log("languageChanged", c))
          : (this.isLanguageChangingTo = void 0),
          s.resolve(function () {
            return r.t(...arguments);
          }),
          e &&
            e(a, function () {
              return r.t(...arguments);
            }));
      },
      l = (a) => {
        !t && !a && this.services.languageDetector && (a = []);
        const c =
          typeof a == "string"
            ? a
            : this.services.languageUtils.getBestMatchFromCodes(a);
        (c &&
          (this.language || i(c),
          this.translator.language || this.translator.changeLanguage(c),
          this.services.languageDetector &&
            this.services.languageDetector.cacheUserLanguage &&
            this.services.languageDetector.cacheUserLanguage(c)),
          this.loadResources(c, (m) => {
            o(m, c);
          }));
      };
    return (
      !t &&
      this.services.languageDetector &&
      !this.services.languageDetector.async
        ? l(this.services.languageDetector.detect())
        : !t &&
            this.services.languageDetector &&
            this.services.languageDetector.async
          ? this.services.languageDetector.detect.length === 0
            ? this.services.languageDetector.detect().then(l)
            : this.services.languageDetector.detect(l)
          : l(t),
      s
    );
  }
  getFixedT(t, e, r) {
    var s = this;
    const i = function (o, l) {
      let a;
      if (typeof l != "object") {
        for (
          var c = arguments.length, m = new Array(c > 2 ? c - 2 : 0), v = 2;
          v < c;
          v++
        )
          m[v - 2] = arguments[v];
        a = s.options.overloadTranslationOptionHandler([o, l].concat(m));
      } else a = { ...l };
      ((a.lng = a.lng || i.lng),
        (a.lngs = a.lngs || i.lngs),
        (a.ns = a.ns || i.ns),
        (a.keyPrefix = a.keyPrefix || r || i.keyPrefix));
      const d = s.options.keySeparator || ".";
      let w;
      return (
        a.keyPrefix && Array.isArray(o)
          ? (w = o.map((E) => `${a.keyPrefix}${d}${E}`))
          : (w = a.keyPrefix ? `${a.keyPrefix}${d}${o}` : o),
        s.t(w, a)
      );
    };
    return (
      typeof t == "string" ? (i.lng = t) : (i.lngs = t),
      (i.ns = e),
      (i.keyPrefix = r),
      i
    );
  }
  t() {
    return this.translator && this.translator.translate(...arguments);
  }
  exists() {
    return this.translator && this.translator.exists(...arguments);
  }
  setDefaultNamespace(t) {
    this.options.defaultNS = t;
  }
  hasLoadedNamespace(t) {
    let e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
    if (!this.isInitialized)
      return (
        this.logger.warn(
          "hasLoadedNamespace: i18next was not initialized",
          this.languages,
        ),
        !1
      );
    if (!this.languages || !this.languages.length)
      return (
        this.logger.warn(
          "hasLoadedNamespace: i18n.languages were undefined or empty",
          this.languages,
        ),
        !1
      );
    const r = e.lng || this.resolvedLanguage || this.languages[0],
      s = this.options ? this.options.fallbackLng : !1,
      i = this.languages[this.languages.length - 1];
    if (r.toLowerCase() === "cimode") return !0;
    const o = (l, a) => {
      const c = this.services.backendConnector.state[`${l}|${a}`];
      return c === -1 || c === 2;
    };
    if (e.precheck) {
      const l = e.precheck(this, o);
      if (l !== void 0) return l;
    }
    return !!(
      this.hasResourceBundle(r, t) ||
      !this.services.backendConnector.backend ||
      (this.options.resources && !this.options.partialBundledLanguages) ||
      (o(r, t) && (!s || o(i, t)))
    );
  }
  loadNamespaces(t, e) {
    const r = ae();
    return this.options.ns
      ? (typeof t == "string" && (t = [t]),
        t.forEach((s) => {
          this.options.ns.indexOf(s) < 0 && this.options.ns.push(s);
        }),
        this.loadResources((s) => {
          (r.resolve(), e && e(s));
        }),
        r)
      : (e && e(), Promise.resolve());
  }
  loadLanguages(t, e) {
    const r = ae();
    typeof t == "string" && (t = [t]);
    const s = this.options.preload || [],
      i = t.filter((o) => s.indexOf(o) < 0);
    return i.length
      ? ((this.options.preload = s.concat(i)),
        this.loadResources((o) => {
          (r.resolve(), e && e(o));
        }),
        r)
      : (e && e(), Promise.resolve());
  }
  dir(t) {
    if (
      (t ||
        (t =
          this.resolvedLanguage ||
          (this.languages && this.languages.length > 0
            ? this.languages[0]
            : this.language)),
      !t)
    )
      return "rtl";
    const e = [
        "ar",
        "shu",
        "sqr",
        "ssh",
        "xaa",
        "yhd",
        "yud",
        "aao",
        "abh",
        "abv",
        "acm",
        "acq",
        "acw",
        "acx",
        "acy",
        "adf",
        "ads",
        "aeb",
        "aec",
        "afb",
        "ajp",
        "apc",
        "apd",
        "arb",
        "arq",
        "ars",
        "ary",
        "arz",
        "auz",
        "avl",
        "ayh",
        "ayl",
        "ayn",
        "ayp",
        "bbz",
        "pga",
        "he",
        "iw",
        "ps",
        "pbt",
        "pbu",
        "pst",
        "prp",
        "prd",
        "ug",
        "ur",
        "ydd",
        "yds",
        "yih",
        "ji",
        "yi",
        "hbo",
        "men",
        "xmn",
        "fa",
        "jpr",
        "peo",
        "pes",
        "prs",
        "dv",
        "sam",
        "ckb",
      ],
      r = (this.services && this.services.languageUtils) || new qe(Je());
    return e.indexOf(r.getLanguagePartFromCode(t)) > -1 ||
      t.toLowerCase().indexOf("-arab") > 1
      ? "rtl"
      : "ltr";
  }
  static createInstance() {
    let t = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {},
      e = arguments.length > 1 ? arguments[1] : void 0;
    return new fe(t, e);
  }
  cloneInstance() {
    let t = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {},
      e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : xe;
    const r = t.forkResourceStore;
    r && delete t.forkResourceStore;
    const s = { ...this.options, ...t, isClone: !0 },
      i = new fe(s);
    return (
      (t.debug !== void 0 || t.prefix !== void 0) &&
        (i.logger = i.logger.clone(t)),
      ["store", "services", "language"].forEach((l) => {
        i[l] = this[l];
      }),
      (i.services = { ...this.services }),
      (i.services.utils = { hasLoadedNamespace: i.hasLoadedNamespace.bind(i) }),
      r &&
        ((i.store = new Ye(this.store.data, s)),
        (i.services.resourceStore = i.store)),
      (i.translator = new we(i.services, s)),
      i.translator.on("*", function (l) {
        for (
          var a = arguments.length, c = new Array(a > 1 ? a - 1 : 0), m = 1;
          m < a;
          m++
        )
          c[m - 1] = arguments[m];
        i.emit(l, ...c);
      }),
      i.init(s, e),
      (i.translator.options = s),
      (i.translator.backendConnector.services.utils = {
        hasLoadedNamespace: i.hasLoadedNamespace.bind(i),
      }),
      i
    );
  }
  toJSON() {
    return {
      options: this.options,
      store: this.store,
      language: this.language,
      languages: this.languages,
      resolvedLanguage: this.resolvedLanguage,
    };
  }
}
const B = fe.createInstance();
B.createInstance = fe.createInstance;
B.createInstance;
B.dir;
B.init;
B.loadResources;
B.reloadResources;
B.use;
B.changeLanguage;
B.getFixedT;
const Ps = B.t;
B.exists;
B.setDefaultNamespace;
B.hasLoadedNamespace;
B.loadNamespaces;
B.loadLanguages;
var Nn = function (t) {
  return {
    type: "backend",
    init: function (r, s, i) {},
    read: function (r, s, i) {
      if (typeof t == "function") {
        if (t.length < 3) {
          try {
            var o = t(r, s);
            o && typeof o.then == "function"
              ? o
                  .then(function (l) {
                    return i(null, (l && l.default) || l);
                  })
                  .catch(i)
              : i(null, o);
          } catch (l) {
            i(l);
          }
          return;
        }
        t(r, s, i);
        return;
      }
      i(null, t && t[r] && t[r][s]);
    },
  };
};
function Dn(n, t) {
  if (!(n instanceof t))
    throw new TypeError("Cannot call a class as a function");
}
function de(n) {
  "@babel/helpers - typeof";
  return (
    (de =
      typeof Symbol == "function" && typeof Symbol.iterator == "symbol"
        ? function (t) {
            return typeof t;
          }
        : function (t) {
            return t &&
              typeof Symbol == "function" &&
              t.constructor === Symbol &&
              t !== Symbol.prototype
              ? "symbol"
              : typeof t;
          }),
    de(n)
  );
}
function Wn(n, t) {
  if (de(n) != "object" || !n) return n;
  var e = n[Symbol.toPrimitive];
  if (e !== void 0) {
    var r = e.call(n, t);
    if (de(r) != "object") return r;
    throw new TypeError("@@toPrimitive must return a primitive value.");
  }
  return String(n);
}
function Mn(n) {
  var t = Wn(n, "string");
  return de(t) == "symbol" ? t : t + "";
}
function Vn(n, t) {
  for (var e = 0; e < t.length; e++) {
    var r = t[e];
    ((r.enumerable = r.enumerable || !1),
      (r.configurable = !0),
      "value" in r && (r.writable = !0),
      Object.defineProperty(n, Mn(r.key), r));
  }
}
function jn(n, t, e) {
  return (
    t && Vn(n.prototype, t),
    Object.defineProperty(n, "prototype", { writable: !1 }),
    n
  );
}
var mt = [],
  Un = mt.forEach,
  $n = mt.slice;
function Bn(n) {
  return (
    Un.call($n.call(arguments, 1), function (t) {
      if (t) for (var e in t) n[e] === void 0 && (n[e] = t[e]);
    }),
    n
  );
}
var Ze = /^[\u0009\u0020-\u007e\u0080-\u00ff]+$/,
  zn = function (t, e, r) {
    var s = r || {};
    s.path = s.path || "/";
    var i = encodeURIComponent(e),
      o = "".concat(t, "=").concat(i);
    if (s.maxAge > 0) {
      var l = s.maxAge - 0;
      if (Number.isNaN(l)) throw new Error("maxAge should be a Number");
      o += "; Max-Age=".concat(Math.floor(l));
    }
    if (s.domain) {
      if (!Ze.test(s.domain)) throw new TypeError("option domain is invalid");
      o += "; Domain=".concat(s.domain);
    }
    if (s.path) {
      if (!Ze.test(s.path)) throw new TypeError("option path is invalid");
      o += "; Path=".concat(s.path);
    }
    if (s.expires) {
      if (typeof s.expires.toUTCString != "function")
        throw new TypeError("option expires is invalid");
      o += "; Expires=".concat(s.expires.toUTCString());
    }
    if (
      (s.httpOnly && (o += "; HttpOnly"),
      s.secure && (o += "; Secure"),
      s.sameSite)
    ) {
      var a =
        typeof s.sameSite == "string" ? s.sameSite.toLowerCase() : s.sameSite;
      switch (a) {
        case !0:
          o += "; SameSite=Strict";
          break;
        case "lax":
          o += "; SameSite=Lax";
          break;
        case "strict":
          o += "; SameSite=Strict";
          break;
        case "none":
          o += "; SameSite=None";
          break;
        default:
          throw new TypeError("option sameSite is invalid");
      }
    }
    return o;
  },
  et = {
    create: function (t, e, r, s) {
      var i =
        arguments.length > 4 && arguments[4] !== void 0
          ? arguments[4]
          : { path: "/", sameSite: "strict" };
      (r &&
        ((i.expires = new Date()),
        i.expires.setTime(i.expires.getTime() + r * 60 * 1e3)),
        s && (i.domain = s),
        (document.cookie = zn(t, encodeURIComponent(e), i)));
    },
    read: function (t) {
      for (
        var e = "".concat(t, "="), r = document.cookie.split(";"), s = 0;
        s < r.length;
        s++
      ) {
        for (var i = r[s]; i.charAt(0) === " "; ) i = i.substring(1, i.length);
        if (i.indexOf(e) === 0) return i.substring(e.length, i.length);
      }
      return null;
    },
    remove: function (t) {
      this.create(t, "", -1);
    },
  },
  Hn = {
    name: "cookie",
    lookup: function (t) {
      var e;
      if (t.lookupCookie && typeof document < "u") {
        var r = et.read(t.lookupCookie);
        r && (e = r);
      }
      return e;
    },
    cacheUserLanguage: function (t, e) {
      e.lookupCookie &&
        typeof document < "u" &&
        et.create(
          e.lookupCookie,
          t,
          e.cookieMinutes,
          e.cookieDomain,
          e.cookieOptions,
        );
    },
  },
  Yn = {
    name: "querystring",
    lookup: function (t) {
      var e;
      if (typeof window < "u") {
        var r = window.location.search;
        !window.location.search &&
          window.location.hash &&
          window.location.hash.indexOf("?") > -1 &&
          (r = window.location.hash.substring(
            window.location.hash.indexOf("?"),
          ));
        for (
          var s = r.substring(1), i = s.split("&"), o = 0;
          o < i.length;
          o++
        ) {
          var l = i[o].indexOf("=");
          if (l > 0) {
            var a = i[o].substring(0, l);
            a === t.lookupQuerystring && (e = i[o].substring(l + 1));
          }
        }
      }
      return e;
    },
  },
  le = null,
  tt = function () {
    if (le !== null) return le;
    try {
      le = window !== "undefined" && window.localStorage !== null;
      var t = "i18next.translate.boo";
      (window.localStorage.setItem(t, "foo"),
        window.localStorage.removeItem(t));
    } catch {
      le = !1;
    }
    return le;
  },
  Kn = {
    name: "localStorage",
    lookup: function (t) {
      var e;
      if (t.lookupLocalStorage && tt()) {
        var r = window.localStorage.getItem(t.lookupLocalStorage);
        r && (e = r);
      }
      return e;
    },
    cacheUserLanguage: function (t, e) {
      e.lookupLocalStorage &&
        tt() &&
        window.localStorage.setItem(e.lookupLocalStorage, t);
    },
  },
  ue = null,
  rt = function () {
    if (ue !== null) return ue;
    try {
      ue = window !== "undefined" && window.sessionStorage !== null;
      var t = "i18next.translate.boo";
      (window.sessionStorage.setItem(t, "foo"),
        window.sessionStorage.removeItem(t));
    } catch {
      ue = !1;
    }
    return ue;
  },
  qn = {
    name: "sessionStorage",
    lookup: function (t) {
      var e;
      if (t.lookupSessionStorage && rt()) {
        var r = window.sessionStorage.getItem(t.lookupSessionStorage);
        r && (e = r);
      }
      return e;
    },
    cacheUserLanguage: function (t, e) {
      e.lookupSessionStorage &&
        rt() &&
        window.sessionStorage.setItem(e.lookupSessionStorage, t);
    },
  },
  Qn = {
    name: "navigator",
    lookup: function (t) {
      var e = [];
      if (typeof navigator < "u") {
        if (navigator.languages)
          for (var r = 0; r < navigator.languages.length; r++)
            e.push(navigator.languages[r]);
        (navigator.userLanguage && e.push(navigator.userLanguage),
          navigator.language && e.push(navigator.language));
      }
      return e.length > 0 ? e : void 0;
    },
  },
  Gn = {
    name: "htmlTag",
    lookup: function (t) {
      var e,
        r =
          t.htmlTag ||
          (typeof document < "u" ? document.documentElement : null);
      return (
        r &&
          typeof r.getAttribute == "function" &&
          (e = r.getAttribute("lang")),
        e
      );
    },
  },
  Jn = {
    name: "path",
    lookup: function (t) {
      var e;
      if (typeof window < "u") {
        var r = window.location.pathname.match(/\/([a-zA-Z-]*)/g);
        if (r instanceof Array)
          if (typeof t.lookupFromPathIndex == "number") {
            if (typeof r[t.lookupFromPathIndex] != "string") return;
            e = r[t.lookupFromPathIndex].replace("/", "");
          } else e = r[0].replace("/", "");
      }
      return e;
    },
  },
  Xn = {
    name: "subdomain",
    lookup: function (t) {
      var e =
          typeof t.lookupFromSubdomainIndex == "number"
            ? t.lookupFromSubdomainIndex + 1
            : 1,
        r =
          typeof window < "u" &&
          window.location &&
          window.location.hostname &&
          window.location.hostname.match(
            /^(\w{2,5})\.(([a-z0-9-]{1,63}\.[a-z]{2,6})|localhost)/i,
          );
      if (r) return r[e];
    },
  };
function Zn() {
  return {
    order: [
      "querystring",
      "cookie",
      "localStorage",
      "sessionStorage",
      "navigator",
      "htmlTag",
    ],
    lookupQuerystring: "lng",
    lookupCookie: "i18next",
    lookupLocalStorage: "i18nextLng",
    lookupSessionStorage: "i18nextLng",
    caches: ["localStorage"],
    excludeCacheFor: ["cimode"],
    convertDetectedLanguage: function (t) {
      return t;
    },
  };
}
var xt = (function () {
  function n(t) {
    var e = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
    (Dn(this, n),
      (this.type = "languageDetector"),
      (this.detectors = {}),
      this.init(t, e));
  }
  return (
    jn(n, [
      {
        key: "init",
        value: function (e) {
          var r =
              arguments.length > 1 && arguments[1] !== void 0
                ? arguments[1]
                : {},
            s =
              arguments.length > 2 && arguments[2] !== void 0
                ? arguments[2]
                : {};
          ((this.services = e || { languageUtils: {} }),
            (this.options = Bn(r, this.options || {}, Zn())),
            typeof this.options.convertDetectedLanguage == "string" &&
              this.options.convertDetectedLanguage.indexOf("15897") > -1 &&
              (this.options.convertDetectedLanguage = function (i) {
                return i.replace("-", "_");
              }),
            this.options.lookupFromUrlIndex &&
              (this.options.lookupFromPathIndex =
                this.options.lookupFromUrlIndex),
            (this.i18nOptions = s),
            this.addDetector(Hn),
            this.addDetector(Yn),
            this.addDetector(Kn),
            this.addDetector(qn),
            this.addDetector(Qn),
            this.addDetector(Gn),
            this.addDetector(Jn),
            this.addDetector(Xn));
        },
      },
      {
        key: "addDetector",
        value: function (e) {
          this.detectors[e.name] = e;
        },
      },
      {
        key: "detect",
        value: function (e) {
          var r = this;
          e || (e = this.options.order);
          var s = [];
          return (
            e.forEach(function (i) {
              if (r.detectors[i]) {
                var o = r.detectors[i].lookup(r.options);
                (o && typeof o == "string" && (o = [o]),
                  o && (s = s.concat(o)));
              }
            }),
            (s = s.map(function (i) {
              return r.options.convertDetectedLanguage(i);
            })),
            this.services.languageUtils.getBestMatchFromCodes
              ? s
              : s.length > 0
                ? s[0]
                : null
          );
        },
      },
      {
        key: "cacheUserLanguage",
        value: function (e, r) {
          var s = this;
          (r || (r = this.options.caches),
            r &&
              ((this.options.excludeCacheFor &&
                this.options.excludeCacheFor.indexOf(e) > -1) ||
                r.forEach(function (i) {
                  s.detectors[i] &&
                    s.detectors[i].cacheUserLanguage(e, s.options);
                })));
        },
      },
    ]),
    n
  );
})();
xt.type = "languageDetector";
const es = (n) => {
    const t = k(n);
    return (
      n.on("initialized", () => {
        t.set(n);
      }),
      n.on("loaded", () => {
        t.set(n);
      }),
      n.on("added", () => t.set(n)),
      n.on("languageChanged", () => {
        t.set(n);
      }),
      t
    );
  },
  ts = (n) => {
    const t = k(!1);
    return (
      n.on("loaded", (e) => {
        Object.keys(e).length !== 0 && t.set(!1);
      }),
      n.on("failedLoading", () => {
        t.set(!0);
      }),
      t
    );
  },
  As = (n) => {
    const t = n
        ? ["querystring", "localStorage"]
        : ["querystring", "localStorage", "navigator"],
      e = n ? [n] : ["en-US"],
      r = (i, o) =>
        fn(
          Object.assign({
            "./locales/ar-BH/translation.json": () =>
              C(() => import("./DCa5p2gA.js"), [], import.meta.url),
            "./locales/ar/translation.json": () =>
              C(() => import("./rpbn34F3.js"), [], import.meta.url),
            "./locales/bg-BG/translation.json": () =>
              C(() => import("./DABO0Imk.js"), [], import.meta.url),
            "./locales/bn-BD/translation.json": () =>
              C(() => import("./DKoD9sRV.js"), [], import.meta.url),
            "./locales/bo-TB/translation.json": () =>
              C(() => import("./CtpSdbCt.js"), [], import.meta.url),
            "./locales/ca-ES/translation.json": () =>
              C(() => import("./DzCUSzGt.js"), [], import.meta.url),
            "./locales/ceb-PH/translation.json": () =>
              C(() => import("./Dr-_J1XK.js"), [], import.meta.url),
            "./locales/cs-CZ/translation.json": () =>
              C(() => import("./CoE-gIuA.js"), [], import.meta.url),
            "./locales/da-DK/translation.json": () =>
              C(() => import("./j6rj1DXq.js"), [], import.meta.url),
            "./locales/de-DE/translation.json": () =>
              C(() => import("./CKpb-rUA.js"), [], import.meta.url),
            "./locales/dg-DG/translation.json": () =>
              C(() => import("./CVqSNjfj.js"), [], import.meta.url),
            "./locales/el-GR/translation.json": () =>
              C(() => import("./DhO_E7PA.js"), [], import.meta.url),
            "./locales/en-GB/translation.json": () =>
              C(() => import("./B6y18JvF.js"), [], import.meta.url),
            "./locales/en-US/mcp_servers.json": () =>
              C(() => import("./x2NOblEu.js"), [], import.meta.url),
            "./locales/en-US/mcp_tools.json": () =>
              C(() => import("./CblF70w-.js"), [], import.meta.url),
            "./locales/en-US/translation.json": () =>
              C(() => import("./DWYiGO-a.js"), [], import.meta.url),
            "./locales/es-ES/translation.json": () =>
              C(() => import("./D9uUIj_Y.js"), [], import.meta.url),
            "./locales/et-EE/translation.json": () =>
              C(() => import("./DCOd3w2N.js"), [], import.meta.url),
            "./locales/eu-ES/translation.json": () =>
              C(() => import("./D7F4qhZn.js"), [], import.meta.url),
            "./locales/fa-IR/translation.json": () =>
              C(() => import("./Biw99-mA.js"), [], import.meta.url),
            "./locales/fi-FI/translation.json": () =>
              C(() => import("./DZ-M0zYD.js"), [], import.meta.url),
            "./locales/fr-CA/translation.json": () =>
              C(() => import("./SDvp2M_8.js"), [], import.meta.url),
            "./locales/fr-FR/translation.json": () =>
              C(() => import("./Caf9roPd.js"), [], import.meta.url),
            "./locales/gl-ES/translation.json": () =>
              C(() => import("./DdbHbupM.js"), [], import.meta.url),
            "./locales/he-IL/translation.json": () =>
              C(() => import("./DK5mvc8B.js"), [], import.meta.url),
            "./locales/hi-IN/translation.json": () =>
              C(() => import("./D_gf84iZ.js"), [], import.meta.url),
            "./locales/hr-HR/translation.json": () =>
              C(() => import("./BB7GebNv.js"), [], import.meta.url),
            "./locales/hu-HU/translation.json": () =>
              C(() => import("./Cr-AU7pL.js"), [], import.meta.url),
            "./locales/id-ID/translation.json": () =>
              C(() => import("./Cx6X7gRt.js"), [], import.meta.url),
            "./locales/ie-GA/translation.json": () =>
              C(() => import("./CIWExUOW.js"), [], import.meta.url),
            "./locales/it-IT/translation.json": () =>
              C(() => import("./C_vmqt5n.js"), [], import.meta.url),
            "./locales/ja-JP/translation.json": () =>
              C(() => import("./DFrgkTjZ.js"), [], import.meta.url),
            "./locales/ka-GE/translation.json": () =>
              C(() => import("./mXA61IDS.js"), [], import.meta.url),
            "./locales/ko-KR/translation.json": () =>
              C(() => import("./DyjuGH0K.js"), [], import.meta.url),
            "./locales/lt-LT/translation.json": () =>
              C(() => import("./BrSwmj6_.js"), [], import.meta.url),
            "./locales/ms-MY/translation.json": () =>
              C(() => import("./ICBPh4EE.js"), [], import.meta.url),
            "./locales/nb-NO/translation.json": () =>
              C(() => import("./81N7lGMv.js"), [], import.meta.url),
            "./locales/nl-NL/translation.json": () =>
              C(() => import("./DLP-auby.js"), [], import.meta.url),
            "./locales/pa-IN/translation.json": () =>
              C(() => import("./BjsS-0_M.js"), [], import.meta.url),
            "./locales/pl-PL/translation.json": () =>
              C(() => import("./BIs-Msyg.js"), [], import.meta.url),
            "./locales/pt-BR/translation.json": () =>
              C(() => import("./B_Nve7Ds.js"), [], import.meta.url),
            "./locales/pt-PT/translation.json": () =>
              C(() => import("./DdNd0L_M.js"), [], import.meta.url),
            "./locales/ro-RO/translation.json": () =>
              C(() => import("./BA3DJ3N-.js"), [], import.meta.url),
            "./locales/ru-RU/translation.json": () =>
              C(() => import("./D5futXdy.js"), [], import.meta.url),
            "./locales/sk-SK/translation.json": () =>
              C(() => import("./Cm1tMdG4.js"), [], import.meta.url),
            "./locales/sr-RS/translation.json": () =>
              C(() => import("./4JK_71iV.js"), [], import.meta.url),
            "./locales/sv-SE/translation.json": () =>
              C(() => import("./62FmMFV0.js"), [], import.meta.url),
            "./locales/th-TH/translation.json": () =>
              C(() => import("./BdkQiw9O.js"), [], import.meta.url),
            "./locales/tk-TM/translation.json": () =>
              C(() => import("./CH5O9Nip.js"), [], import.meta.url),
            "./locales/tk-TW/translation.json": () =>
              C(() => import("./Cfvu6c3o.js"), [], import.meta.url),
            "./locales/tr-TR/translation.json": () =>
              C(() => import("./D-3rZByG.js"), [], import.meta.url),
            "./locales/uk-UA/translation.json": () =>
              C(() => import("./DLFOR3DW.js"), [], import.meta.url),
            "./locales/ur-PK/translation.json": () =>
              C(() => import("./D4M-3pz1.js"), [], import.meta.url),
            "./locales/vi-VN/translation.json": () =>
              C(() => import("./BGGW79PL.js"), [], import.meta.url),
            "./locales/zh-CN/mcp_servers.json": () =>
              C(() => import("./_9sXXz4r.js"), [], import.meta.url),
            "./locales/zh-CN/mcp_tools.json": () =>
              C(() => import("./ChO4R-Xl.js"), [], import.meta.url),
            "./locales/zh-CN/translation.json": () =>
              C(() => import("./I3Oqw-P8.js"), [], import.meta.url),
            "./locales/zh-TW/translation.json": () =>
              C(() => import("./D7iRnvrv.js"), [], import.meta.url),
          }),
          `./locales/${i === "zh" ? "zh-CN" : i === "en" ? "en-US" : i}/${o}.json`,
          4,
        );
    B.use(Nn(r))
      .use(xt)
      .init({
        debug: !1,
        detection: {
          order: t,
          caches: ["localStorage"],
          lookupQuerystring: "lang",
          lookupLocalStorage: "locale",
        },
        fallbackLng: {
          default: e,
          mcp_tools: ["en-US", ...e],
          mcp_servers: ["en-US", ...e],
        },
        ns: ["translation", "mcp_tools", "mcp_servers"],
        defaultNS: "translation",
        returnEmptyString: !1,
        interpolation: { escapeValue: !1 },
      });
    const s = (B == null ? void 0 : B.language) || n || "en-US";
    (wt.set(s), document.documentElement.setAttribute("lang", s));
  },
  Ie = es(B);
ts(B);
const Fs = (n) => {
  (document.documentElement.setAttribute("lang", n),
    B.changeLanguage(n, (t) => {
      t && console.error("changeLanguage error:", t);
    }),
    wt.set(n));
};
(function (n, t) {
  const e = {
      _0x20a336: 352,
      _0x109fbe: "qukk",
      _0x30b7ba: "FLqM",
      _0x2c1899: "oC#U",
      _0x4b987a: 357,
      _0x3abbd6: "ESb7",
      _0x433d62: 295,
      _0xb268e2: "Kr&$",
      _0x378628: 308,
      _0x13af98: "^O#o",
      _0x36ef80: 359,
      _0x332b5f: 318,
      _0x23e24a: "2Y)w",
      _0x4a9107: 287,
    },
    r = K,
    s = n();
  for (;;)
    try {
      if (
        (-parseInt(r(e._0x20a336, e._0x109fbe)) / 1) *
          (-parseInt(r(327, e._0x30b7ba)) / 2) +
          parseInt(r(335, e._0x2c1899)) / 3 +
          (parseInt(r(e._0x4b987a, e._0x3abbd6)) / 4) *
            (parseInt(r(307, "^O#o")) / 5) +
          (-parseInt(r(e._0x433d62, e._0xb268e2)) / 6) *
            (-parseInt(r(323, "0Cgs")) / 7) +
          (-parseInt(r(e._0x378628, "L3ap")) / 8) *
            (parseInt(r(300, e._0x13af98)) / 9) +
          -parseInt(r(e._0x36ef80, "L3ap")) / 10 +
          (parseInt(r(e._0x332b5f, e._0x23e24a)) / 11) *
            (-parseInt(r(e._0x4a9107, "7Ev4")) / 12) ===
        t
      )
        break;
      s.push(s.shift());
    } catch {
      s.push(s.shift());
    }
})(Se, 436420);
function K(n, t) {
  const e = Se();
  return (
    (K = function (r, s) {
      r = r - 284;
      let i = e[r];
      if (K.Ohqbga === void 0) {
        var o = function (v) {
          const d =
            "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/=";
          let w = "",
            E = "";
          for (
            let p = 0, x, f, h = 0;
            (f = v.charAt(h++));
            ~f && ((x = p % 4 ? x * 64 + f : f), p++ % 4)
              ? (w += String.fromCharCode(255 & (x >> ((-2 * p) & 6))))
              : 0
          )
            f = d.indexOf(f);
          for (let p = 0, x = w.length; p < x; p++)
            E += "%" + ("00" + w.charCodeAt(p).toString(16)).slice(-2);
          return decodeURIComponent(E);
        };
        const m = function (v, d) {
          let w = [],
            E = 0,
            p,
            x = "";
          v = o(v);
          let f;
          for (f = 0; f < 256; f++) w[f] = f;
          for (f = 0; f < 256; f++)
            ((E = (E + w[f] + d.charCodeAt(f % d.length)) % 256),
              (p = w[f]),
              (w[f] = w[E]),
              (w[E] = p));
          ((f = 0), (E = 0));
          for (let h = 0; h < v.length; h++)
            ((f = (f + 1) % 256),
              (E = (E + w[f]) % 256),
              (p = w[f]),
              (w[f] = w[E]),
              (w[E] = p),
              (x += String.fromCharCode(
                v.charCodeAt(h) ^ w[(w[f] + w[E]) % 256],
              )));
          return x;
        };
        ((K.INxVCP = m), (n = arguments), (K.Ohqbga = !0));
      }
      const l = e[0],
        a = r + l,
        c = n[a];
      return (
        c
          ? (i = c)
          : (K.NEogcC === void 0 && (K.NEogcC = !0),
            (i = K.INxVCP(i, s)),
            (n[a] = i)),
        i
      );
    }),
    K(n, t)
  );
}
function Se() {
  const n = [
    "tmkfW4qDWR/dUsbNEfxdJe8",
    "b8kFwu3dOtTGWQxcPG",
    "E3hdGSkrW7j+W5ZcJa",
    "mHTIgx9V",
    "WO5KW7Hjx8o1nuq",
    "WOGSWQZdPG",
    "W7PVnSkJrSkHyCkbW48",
    "WOP4CCo5WRNdQLldKa",
    "W6ndWP5aWRxcLIaIbCotWR7dR8oU",
    "WOxdSmogD0eAoHy",
    "ySkrW5NcOYBcU8oP",
    "fKFdO8kxW77cOComauij",
    "WO0LW7ZcQmouW5RdUSo/BCktd8ke",
    "DCkZWQ8SdCou",
    "uCoGuqtcSeFcP8kJ",
    "lSkWvdLL",
    "dmoEWOjlW6lcPq",
    "xmoeeNxdUa1jWOC",
    "WQ8osmkZWRXknSoK",
    "e8oyWP5jW7pcOW",
    "kSk6qIHOeq",
    "W6FdMCokgCoKWQxcO8oepCkiAG",
    "jNVcONlcSG",
    "fKFdO8kxW77cVSoadfejWQS",
    "WQCxWPFcQcVcNLtcSuJcS8ogdG",
    "nCkIWOLrF8k0W4m",
    "W6FdMCoOl8of",
    "rmoSpmkKWQxcTG",
    "fv3dShRcImoaWRmgmmom",
    "WQn8W5VdOaBdHCoyECkBaCkFyG",
    "yCkjDrrJkWC",
    "WO7dMSkcWQdcKahdR8oPdCkkkq",
    "WQv0W53dOqRdHCkmz8kznCkdxfG",
    "cmouWPrAW68",
    "tmoHBbxcShNcOmkGmmkD",
    "ymk3WQfXsG",
    "W7FcVmkaW5bgW7X/vG",
    "WO7dI8oSoZb5",
    "wIOGWQK+p8ojl8kRWOhcQCopvW",
    "W4bxWOpcMq",
    "W4xdGmoHAflcR8kCeCoLqmo0zW",
    "fSkYBCoZW7NdP8ouW4ONW50qW4ldLq",
    "WP/cQCo9kfFcVsFdPCkOheSRFW",
    "WPFdHCoJWQu0",
    "W4RdVmk9CGZdQehdNSkKf2mFDaxcVfK",
    "uCkTWPbxW61vW5myW7xdVSoSkmk3xhi",
    "W49WuXtdVmkVCu7cKSk8W78",
    "v8kNWRbmW7nkW5GB",
    "tX3dP21EsCosW7dcHa",
    "W6ZcTSkb",
    "sSoAWRVcRa",
    "BM3dMCkxW7r4W4a",
    "smoUEItcRvVcQSkSfmkAymo1W5NdPa",
    "WR1fBb8kWOLNW793nXa",
    "W6eRWP3cRq",
    "WO56C8o7",
    "W6ddLCoXl8ooWPG",
    "W7fbW5FdQxRdIv8",
    "i8kLmSoIW5bfg8oPW7CJW4S+ca",
    "nCk2uYX5fIPL",
    "WRXgBHOcW7PWW5LdnIlcKa",
    "WQf1W53cLulcPSoWsmkz",
    "sCoUBbFcT0/cRSkHnW",
    "WOe4W7RcQW",
    "WOu4W6q",
    "WRLdAX8eWRDNW5LojHG",
    "smkTWPOvW4fJW7y8WPpcP8kXACoXgINdVfzBxCkBxSk7W40XBSk4C8kyW6lcIuC",
    "dvVdNIq",
    "WOOMWQZdR3eA",
    "EsHvjxrvWQS",
    "W5nTyYy",
    "lbTjW7XJuHqN",
    "WPVdUSodC0CbmaO",
    "WOqSWRFdJxGnzW",
    "WPvYC8oQWQ7dQW",
    "cedcQ01m",
    "W4xcOSkxiaqqdrb3c3S",
    "D8o/xKldRs8XWQHJW4ldUsu",
    "pSkWWQpdNSk3WO0CW4y",
    "W6pcISkNW65rBrxcTrrlWR7dNeiD",
  ];
  return (
    (Se = function () {
      return n;
    }),
    Se()
  );
}
const rs = () => Z.isMobile || Z.isTablet,
  Ns = () => {
    var v;
    const n = {
        _0x561ad7: "j9jt",
        _0x5ebd53: 296,
        _0x43c967: "3b@g",
        _0x1e86cc: 353,
        _0x5c8a72: "*OV2",
        _0x596107: 332,
        _0x42ed35: "O*bQ",
        _0xfafbfe: "dXt$",
        _0x1becb7: 285,
        _0x2d7152: "WZq4",
        _0x6ed592: 330,
        _0x4ac283: "XrJd",
        _0x483757: 303,
        _0x48e4dc: "ESb7",
        _0x240e5b: "O*bQ",
        _0x28c3d3: 358,
        _0xb21b6d: 343,
        _0x4913d2: "FLqM",
        _0xad85f5: "y^M1",
        _0xbaff41: 291,
        _0x358e91: "&k33",
        _0x2091d3: 317,
        _0x13932f: "ltFI",
        _0x3595a4: 306,
        _0x10c657: 321,
        _0x305306: "Ty[K",
        _0x300ae0: 286,
        _0x5e3a51: "dG4[",
        _0x33ba99: 314,
        _0x481d99: 344,
        _0x1d4911: "RohJ",
        _0x3b61fb: 334,
        _0x578c11: 329,
        _0x155d53: "mWSl",
        _0x4ce6ff: "0Cgs",
        _0x52a3aa: 322,
        _0x4cef7e: 348,
        _0x3ce4c7: "n9w0",
        _0x53d290: 324,
        _0x238e40: "AnPM",
        _0x54857c: "*OV2",
        _0xdbed1b: 298,
        _0x4cc29f: 320,
        _0x3efba4: 341,
        _0x4da965: "*OV2",
        _0x2ed9cb: 304,
        _0x587013: "0bT3",
        _0xf6840c: 301,
        _0x204cb9: "L3ap",
        _0x568bac: "l3DJ",
      },
      t = { _0x2b6348: 284 },
      e = K,
      r = String(Date[e(311, "l3DJ")]()),
      s = ut(),
      i = ((v = ne(Fe)) == null ? void 0 : v.id) || "",
      o = { timestamp: r, requestId: s, user_id: i },
      l = {
        version: e(290, n._0x561ad7),
        platform: e(n._0x5ebd53, "FrN4"),
        token: localStorage[e(337, n._0x43c967)](e(n._0x1e86cc, "n9w0")) || "",
        user_agent: navigator.userAgent,
        language: navigator.language,
        languages:
          navigator[e(309, n._0x5c8a72)][e(n._0x596107, n._0x42ed35)](","),
        timezone: Intl[e(326, "j9jt")]()[e(292, n._0xfafbfe)]().timeZone,
        cookie_enabled: String(navigator[e(n._0x1becb7, n._0x2d7152)]),
        screen_width: String(
          window[e(n._0x6ed592, n._0x4ac283)][e(342, "ESb7")],
        ),
        screen_height: String(window.screen.height),
        screen_resolution:
          window[e(n._0x483757, "n9w0")][e(360, "FLqM")] +
          "x" +
          window[e(347, n._0x48e4dc)][e(315, n._0x240e5b)],
        viewport_height: String(window[e(n._0x28c3d3, "jY!q")]),
        viewport_width: String(window[e(361, "*OV2")]),
        viewport_size: window[e(338, "!PH4")] + "x" + window[e(350, "!PH4")],
        color_depth: String(
          window[e(n._0xb21b6d, n._0x4913d2)][e(355, n._0xad85f5)],
        ),
        pixel_ratio: String(window[e(n._0xbaff41, n._0x358e91)]),
        current_url: window[e(325, "59^9")][e(n._0x2091d3, n._0x13932f)],
        pathname: window[e(n._0x3595a4, n._0x48e4dc)].pathname,
        search: window[e(319, "0Cgs")][e(n._0x10c657, n._0x305306)],
        hash: window.location[e(n._0x300ae0, n._0x5e3a51)],
        host: window[e(363, "FrN4")][e(n._0x33ba99, "qLSN")],
        hostname: window.location[e(n._0x481d99, n._0x1d4911)],
        protocol: window[e(n._0x3b61fb, "Ty[K")][e(n._0x578c11, n._0x155d53)],
        referrer: document[e(336, n._0x4ce6ff)],
        title: document[e(n._0x52a3aa, "Kr&$")],
        timezone_offset: String(new Date().getTimezoneOffset()),
        local_time: new Date()[e(n._0x4cef7e, n._0x3ce4c7)](),
        utc_time: new Date()[e(293, "ltFI")](),
        is_mobile: rs()[e(294, n._0xfafbfe)](),
        is_touch: String(e(n._0x53d290, n._0x238e40) in window),
        max_touch_points: String(navigator[e(299, n._0x54857c)]),
        browser_name: Z.browserName,
        os_name: Z.osName,
      },
      a = { ...o, ...l },
      c = new URLSearchParams();
    Object[e(n._0xdbed1b, "mWSl")](a)[e(n._0x4cc29f, n._0x240e5b)](([d, w]) => {
      c[e(t._0x2b6348, "X&CA")](d, String(w));
    });
    const m = c[e(n._0x3efba4, n._0x4da965)]();
    return {
      sortedPayload: Object[e(n._0x2ed9cb, n._0x587013)](o)
        [e(n._0xf6840c, n._0x204cb9)]((d, w) => d[0][e(305, "#BqI")](w[0]))
        [e(310, n._0x568bac)](","),
      urlParams: m,
      timestamp: r,
    };
  },
  Ds = (n, t, e) => {
    const r = {
        _0x5be688: 354,
        _0x546f53: "jtCr",
        _0x21ca35: 346,
        _0x275322: "FLqM",
        _0x5314db: 339,
        _0x2c3256: "l3DJ",
        _0x4bb8c6: "O2l[",
        _0x4a6bdc: 362,
        _0x544fcd: "qukk",
        _0x8a34bb: 340,
        _0x508afa: 302,
        _0x283abe: "Ty[K",
        _0x272db5: 313,
      },
      s = K,
      i = Number(e),
      o = e,
      l = new TextEncoder(),
      a = l[s(r._0x5be688, r._0x546f53)](t),
      c = 32768;
    let m = "";
    for (let x = 0; x < a[s(r._0x21ca35, r._0x275322)]; x += c) {
      const f = a.slice(x, x + c);
      m += String[s(r._0x5314db, r._0x2c3256)][s(349, r._0x4bb8c6)](
        null,
        Array.from(f),
      );
    }
    const v = btoa(m),
      d = n + "|" + v + "|" + o,
      w = Math[s(r._0x4a6bdc, r._0x544fcd)](i / (5 * 60 * 1e3)),
      E = Be[s(r._0x8a34bb, r._0x544fcd)][s(r._0x508afa, r._0x283abe)](
        s(r._0x272db5, "dXt$"),
        "" + w,
      );
    return {
      signature: Be.sha256[s(297, "xo!l")](E, d)[s(331, "mWW0")](),
      timestamp: o,
    };
  };
se.extend(Kr);
se.extend(on);
se.extend(ln);
se.extend(cn);
const Ws = (n) => new Promise((t) => setTimeout(t, n));
function ns(n) {
  return n.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
const Ms = (n, t, e, r) => {
    const s = [
      { regex: /{{char}}/gi, replacement: e },
      { regex: /{{user}}/gi, replacement: r },
      {
        regex: /{{VIDEO_FILE_ID_([a-f0-9-]+)}}/gi,
        replacement: (o, l) =>
          `<video src="${je}/api/v1/files/${l}/content" controls></video>`,
      },
      {
        regex: /{{HTML_FILE_ID_([a-f0-9-]+)}}/gi,
        replacement: (o, l) =>
          `<iframe src="${je}/api/v1/files/${l}/content/html" width="100%" frameborder="0" onload="this.style.height=(this.contentWindow.document.body.scrollHeight+20)+'px';"></iframe>`,
      },
    ];
    return (
      (n = ((o, l) =>
        o
          .split(/(```[\s\S]*?```|`[\s\S]*?`)/)
          .map((a) => (a.startsWith("```") || a.startsWith("`") ? a : l(a)))
          .join(""))(
        n,
        (o) => (
          s.forEach(({ regex: l, replacement: a }) => {
            a != null && (o = o.replace(l, a));
          }),
          Array.isArray(t) &&
            t.forEach((l, a) => {
              const c = new RegExp(`\\[${a + 1}\\]`, "g");
              o = o.replace(c, `<source_id data="${a + 1}" title="${l}" />`);
            }),
          o
        ),
      )),
      n
    );
  },
  Vs = (n) =>
    n
      .replace(/<\|[a-z]*$/, "")
      .replace(/<\|[a-z]+\|$/, "")
      .replace(/<$/, "")
      .replaceAll(/<\|[a-z]+\|>/g, " ")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .trim();
function ss(n) {
  return n
    .replace(
      /\n\$/g,
      `
 $`,
    )
    .replace(/\$\n/g, "$ ");
}
function is(n) {
  return n.replace(/```html<html/g, "```html\n<html");
}
const js = (n) => ((n = bs(n)), (n = ss(n)), (n = is(n)), n.trim());
function Us(n) {
  return new DOMParser().parseFromString(n, "text/html").documentElement
    .textContent;
}
const $s = (n) => {
    let t = "";
    return new TransformStream({
      transform(e, r) {
        t += e;
        const s = t.split(n);
        (s.slice(0, -1).forEach((i) => r.enqueue(i)), (t = s[s.length - 1]));
      },
      flush(e) {
        t && e.enqueue(t);
      },
    });
  },
  Bs = (n) => {
    const t = { messages: {}, currentId: null };
    let e = null,
      r = null;
    for (const s of n)
      ((r = ut()),
        e !== null &&
          (t.messages[e].childrenIds = [...t.messages[e].childrenIds, r]),
        (t.messages[r] = { ...s, id: r, parentId: e, childrenIds: [] }),
        (e = r));
    return ((t.currentId = r), t);
  },
  zs = (n) => {
    const t = n.lastIndexOf('<details type="reasoning" done="false"');
    return t !== -1 && !n.substring(t).includes("</details>")
      ? (console.error("fix </details>"),
        n +
          `
</details>`)
      : n;
  },
  os = () => {
    const n = document.createElement("canvas"),
      t = n.getContext("2d");
    ((n.height = 1), (n.width = 1));
    const e = new ImageData(n.width, n.height),
      r = e.data;
    for (let i = 0; i < e.data.length; i += 1)
      i % 4 !== 3 ? (r[i] = Math.floor(256 * Math.random())) : (r[i] = 255);
    t.putImageData(e, 0, 0);
    const s = t.getImageData(0, 0, n.width, n.height).data;
    for (let i = 0; i < s.length; i += 1)
      if (s[i] !== r[i])
        return (
          console.log(
            "canvasPixelTest: Wrong canvas pixel RGB value detected:",
            s[i],
            "at:",
            i,
            "expected:",
            r[i],
          ),
          console.log("canvasPixelTest: Canvas blocking or spoofing is likely"),
          !1
        );
    return !0;
  },
  Hs = (n) => {
    if (!n) return "/user.png";
    const t = document.createElement("canvas"),
      e = t.getContext("2d");
    if (((t.width = 100), (t.height = 100), !os()))
      return (
        console.log(
          "generateInitialsImage: failed pixel test, fingerprint evasion is likely. Using default image.",
        ),
        "/user.png"
      );
    ((e.fillStyle = "#F39C12"),
      e.fillRect(0, 0, t.width, t.height),
      (e.fillStyle = "#FFFFFF"),
      (e.font = "40px Helvetica"),
      (e.textAlign = "center"),
      (e.textBaseline = "middle"));
    const r = n.trim(),
      s =
        r.length > 0
          ? r[0] + (r.split(" ").length > 1 ? r[r.lastIndexOf(" ") + 1] : "")
          : "";
    return (
      e.fillText(s.toUpperCase(), t.width / 2, t.height / 2),
      t.toDataURL()
    );
  },
  Ys = (n) => {
    const t = se(n);
    return (
      se(),
      t.isToday()
        ? `Today at ${t.format("LT")}`
        : t.isYesterday()
          ? `Yesterday at ${t.format("LT")}`
          : `${t.format("L")} at ${t.format("LT")}`
    );
  },
  Ks = async (n) => {
    let t = !1;
    if (!navigator.clipboard) {
      const e = document.createElement("textarea");
      ((e.value = n),
        (e.style.top = "0"),
        (e.style.left = "0"),
        (e.style.position = "fixed"),
        document.body.appendChild(e),
        e.focus(),
        e.select());
      try {
        const s = document.execCommand("copy") ? "successful" : "unsuccessful";
        (console.log("Fallback: Copying text command was " + s), (t = !0));
      } catch (r) {
        console.error("Fallback: Oops, unable to copy", r);
      }
      return (document.body.removeChild(e), t);
    }
    return (
      (t = await navigator.clipboard
        .writeText(n)
        .then(
          () => (
            console.log("Async: Copying to clipboard was successful!"),
            !0
          ),
        )
        .catch((e) => (console.error("Async: Could not copy text: ", e), !1))),
      t
    );
  },
  qs = (n, t) =>
    t === "0.0.0"
      ? !1
      : t.localeCompare(n, void 0, {
          numeric: !0,
          sensitivity: "case",
          caseFirst: "upper",
        }) < 0,
  Qs = (n) => {
    const t = /\{\{([^}]+)\}\}/g,
      e = [];
    let r;
    for (; (r = t.exec(n)) !== null; )
      e.push({
        word: r[1].trim(),
        startIndex: r.index,
        endIndex: t.lastIndex - 1,
      });
    return e;
  },
  Gs = (n) => ("mapping" in n[0] ? "openai" : "webui"),
  Js = async (n = !1) => {
    const t = await new Promise((s, i) => {
      navigator.geolocation.getCurrentPosition(s, i);
    }).catch((s) => {
      throw (console.error("Error getting user location:", s), s);
    });
    if (!t) return "Location not available";
    const { latitude: e, longitude: r } = t.coords;
    return n
      ? { latitude: e, longitude: r }
      : `${e.toFixed(3)}, ${r.toFixed(3)} (lat, long)`;
  },
  as = (n) => {
    var l, a, c, m, v, d, w, E;
    const t = n.mapping,
      e = [];
    let r = "",
      s = null;
    for (const p in t) {
      const x = t[p];
      r = p;
      try {
        if (
          e.length == 0 &&
          (x.message == null ||
            (((l = x.message.content.parts) == null ? void 0 : l[0]) == "" &&
              x.message.content.text == null))
        )
          continue;
        {
          const f = {
            id: p,
            parentId: s,
            childrenIds: x.children || [],
            role:
              ((c = (a = x.message) == null ? void 0 : a.author) == null
                ? void 0
                : c.role) !== "user"
                ? "assistant"
                : "user",
            content:
              ((d =
                (v = (m = x.message) == null ? void 0 : m.content) == null
                  ? void 0
                  : v.parts) == null
                ? void 0
                : d[0]) ||
              ((E = (w = x.message) == null ? void 0 : w.content) == null
                ? void 0
                : E.text) ||
              "",
            model: "gpt-3.5-turbo",
            done: !0,
            context: null,
          };
          (e.push(f), (s = r));
        }
      } catch (f) {
        console.log(
          "Error with",
          x,
          `
Error:`,
          f,
        );
      }
    }
    const i = {};
    return (
      e.forEach((p) => (i[p.id] = p)),
      {
        history: { currentId: r, messages: i },
        models: ["gpt-3.5-turbo"],
        messages: e,
        options: {},
        timestamp: n.create_time,
        title: n.title ?? "New Chat",
      }
    );
  },
  ls = (n) => {
    const t = n.messages;
    if (
      t.length === 0 ||
      t[t.length - 1].childrenIds.length !== 0 ||
      t[0].parentId !== null
    )
      return !1;
    for (const s of t) if (typeof s.content != "string") return !1;
    return !0;
  },
  Xs = (n) => {
    const t = [];
    let e = 0;
    for (const r of n) {
      const s = as(r);
      ls(s)
        ? t.push({
            id: r.id,
            user_id: "",
            title: r.title,
            chat: s,
            timestamp: r.create_time,
          })
        : e++;
    }
    return (console.log(e, "Conversations could not be imported"), t);
  },
  Zs = (n, t, e = "") => {
    let r = n;
    for (const s of t)
      r = r.replace(
        new RegExp(
          `(
\\s*)*<details\\s+type="${s}"[^>]*>.*?<\\/details>(
\\s*)*`,
          "gis",
        ),
        e,
      );
    return r;
  },
  ei = (n, t = "") => (
    (n = n.replace(
      new RegExp(
        `(
\\s*)*<glm_block\\s+[^>]*>.*?<\\/glm_block>(
\\s*)*`,
        "gis",
      ),
      t,
    )),
    n
  ),
  ti = (n, t) => ({
    "{{USER_NAME}}": n,
    "{{USER_LOCATION}}": t || "Unknown",
    "{{CURRENT_DATETIME}}": us(),
    "{{CURRENT_DATE}}": bt(),
    "{{CURRENT_TIME}}": vt(),
    "{{CURRENT_WEEKDAY}}": _t(),
    "{{CURRENT_TIMEZONE}}": yt(),
    "{{USER_LANGUAGE}}": localStorage.getItem("locale") || "en-US",
  }),
  ri = (n, t, e) => {
    const r = new Date(),
      s =
        r.getFullYear() +
        "-" +
        String(r.getMonth() + 1).padStart(2, "0") +
        "-" +
        String(r.getDate()).padStart(2, "0"),
      i = r.toLocaleTimeString("en-US", {
        hour: "numeric",
        minute: "numeric",
        second: "numeric",
        hour12: !0,
      }),
      o = _t(),
      l = yt(),
      a = localStorage.getItem("locale") || "en-US";
    return (
      (n = n.replace("{{CURRENT_DATETIME}}", `${s} ${i}`)),
      (n = n.replace("{{CURRENT_DATE}}", s)),
      (n = n.replace("{{CURRENT_TIME}}", i)),
      (n = n.replace("{{CURRENT_WEEKDAY}}", o)),
      (n = n.replace("{{CURRENT_TIMEZONE}}", l)),
      (n = n.replace("{{USER_LANGUAGE}}", a)),
      t && (n = n.replace("{{USER_NAME}}", t)),
      e
        ? (n = n.replace("{{USER_LOCATION}}", e))
        : (n = n.replace("{{USER_LOCATION}}", "LOCATION_UNKNOWN")),
      n
    );
  },
  ni = (n) => {
    const t = new Date(),
      e = new Date(n * 1e3),
      s = (t.getTime() - e.getTime()) / (1e3 * 3600 * 24),
      i = t.getDate(),
      o = t.getMonth(),
      l = t.getFullYear(),
      a = e.getDate(),
      c = e.getMonth(),
      m = e.getFullYear();
    return l === m && o === c && i === a
      ? "Today"
      : l === m && o === c && i - a === 1
        ? "Yesterday"
        : s <= 7
          ? "Previous 7 days"
          : s <= 30
            ? "Previous 30 days"
            : l === m
              ? e.toLocaleString("en-US", { month: "long" })
              : e.getFullYear().toString();
  },
  si = (n) => {
    const t = {};
    let e = !1,
      r = !1;
    const s = /^\s*([a-z_]+):\s*(.*)\s*$/i,
      i = n.split(`
`);
    if (i[0].trim() !== '"""') return {};
    e = !0;
    for (let o = 1; o < i.length; o++) {
      const l = i[o];
      if (l.includes('"""') && e) {
        r = !0;
        break;
      }
      if (e && !r) {
        const a = s.exec(l);
        if (a) {
          const [, c, m] = a;
          t[c.trim()] = m.trim();
        }
      }
    }
    return t;
  },
  bt = () => {
    const n = new Date(),
      t = n.getFullYear(),
      e = String(n.getMonth() + 1).padStart(2, "0"),
      r = String(n.getDate()).padStart(2, "0");
    return `${t}-${e}-${r}`;
  },
  vt = () => new Date().toTimeString().split(" ")[0],
  us = () => `${bt()} ${vt()}`,
  yt = () => Intl.DateTimeFormat().resolvedOptions().timeZone,
  _t = () =>
    [
      "Sunday",
      "Monday",
      "Tuesday",
      "Wednesday",
      "Thursday",
      "Friday",
      "Saturday",
    ][new Date().getDay()],
  cs = (n, t) => {
    if (t === null) return [];
    const e = n.messages[t];
    return e != null && e.parentId ? [...cs(n, e.parentId), e] : [e];
  },
  ii = (n, t) => {
    var s;
    if (!(n != null && n.messages) || Object.keys(n.messages).length === 0)
      return null;
    let e = null,
      r = 0;
    for (const [i, o] of Object.entries(n.messages)) {
      const l = o.timestamp || 0;
      o.model === t &&
        (l > r ||
          (l === r &&
            o.role === "assistant" &&
            ((s = n.messages[e]) == null ? void 0 : s.role) !== "assistant")) &&
        ((r = l), (e = i));
    }
    return e;
  },
  oi = (n) => {
    if (n == null) return "Unknown size";
    if (typeof n != "number" || n < 0) return "Invalid size";
    if (n === 0) return "0 B";
    const t = ["B", "KB", "MB", "GB", "TB"];
    let e = 0;
    for (; n >= 1024 && e < t.length - 1; ) ((n /= 1024), e++);
    return `${n.toFixed(1)} ${t[e]}`;
  },
  ai = (n) => (
    console.log(typeof n),
    n
      ? n.split(`
`).length
      : 0
  );
function li(n) {
  let t = 0,
    e = !1;
  for (let r = 0; r < n.length; r++) {
    const s = n[r],
      i = /[\u4e00-\u9fff]/.test(s),
      o = /^[a-zA-Z0-9]+$/.test(s),
      l = /[.,;!?(){}\[\]<>:"'`~\-+*/&^%$#@|\\]/.test(s);
    i
      ? ((t += 1), e && ((t += 1), (e = !1)))
      : o || l
        ? e || (e = !0)
        : (s === " " ||
            s ===
              `
` ||
            (t += 1),
          e && ((t += 1), (e = !1)));
  }
  return (e && (t += 1), t);
}
function be(n, t, e = new Set()) {
  if (!n) return {};
  if (n.$ref) {
    const s = n.$ref.split("/").pop();
    if (e.has(s)) return {};
    e.add(s);
    const i = t.schemas[s];
    return be(i, t, e);
  }
  if (n.type) {
    const r = { type: n.type };
    switch ((n.description && (r.description = n.description), n.type)) {
      case "object":
        ((r.properties = {}), (r.required = n.required || []));
        for (const [s, i] of Object.entries(n.properties || {}))
          r.properties[s] = be(i, t);
        break;
      case "array":
        r.items = be(n.items, t);
        break;
    }
    return r;
  }
  return {};
}
const ui = (n) => {
    const t = [];
    for (const [e, r] of Object.entries(n.paths))
      for (const [s, i] of Object.entries(r)) {
        const o = {
          type: "function",
          name: i.operationId,
          description:
            i.description || i.summary || "No description available.",
          parameters: { type: "object", properties: {}, required: [] },
        };
        if (
          (i.parameters &&
            i.parameters.forEach((l) => {
              ((o.parameters.properties[l.name] = {
                type: l.schema.type,
                description: l.schema.description || "",
              }),
                l.required && o.parameters.required.push(l.name));
            }),
          i.requestBody)
        ) {
          const l = i.requestBody.content;
          if (l && l["application/json"]) {
            const a = l["application/json"].schema,
              c = be(a, n.components);
            c.properties
              ? ((o.parameters.properties = {
                  ...o.parameters.properties,
                  ...c.properties,
                }),
                c.required &&
                  (o.parameters.required = [
                    ...new Set([...o.parameters.required, ...c.required]),
                  ]))
              : c.type === "array" && (o.parameters = c);
          }
        }
        t.push(o);
      }
    return t;
  },
  ci = (n) => Object.keys(n).length === 0;
function fi(n, t, e = "chain") {
  (t(n), (n = fs(n, e)));
  for (let r = 0; r < n.length; r++) {
    const s = n[r];
    r !== n.length - 1 && (s.done = !0);
  }
  return n;
}
function fs(n, t = "chain") {
  var s, i;
  const e = [];
  let r = 0;
  for (const o of n) {
    if (o.type === "space" && Array.isArray(e[r])) {
      ((o.hidden = !0), e[r].push(o));
      continue;
    }
    if (
      o.type === "details" &&
      ((s = o.attributes) == null ? void 0 : s.type) !== "reasoning" &&
      o.summary
    ) {
      Array.isArray(e[r])
        ? (r++, (e[r] = o))
        : (e[r] && r++, (o.view = "html"), (e[r] = o), r++);
      continue;
    }
    if (o.type !== "glm_block" && o.type !== "details") {
      Array.isArray(e[r]) ? (r++, (e[r] = o)) : (e[r] && r++, (e[r] = o), r++);
      continue;
    }
    if (
      ((i = o.attributes) == null ? void 0 : i.tool_call_name) === "retrieve"
    ) {
      (e[r] && r++, (e[r] = o), r++);
      continue;
    }
    if (t === "chain" && (o.type === "details" || o.type === "glm_block")) {
      Array.isArray(e[r]) ? e[r].push(o) : (e[r] && r++, (e[r] = [o]));
      continue;
    }
    if (o.type === "details") {
      Array.isArray(e[r])
        ? (e[r].push(o), r++)
        : (e[r] && r++, (e[r] = o), r++);
      continue;
    }
    if (o.type === "glm_block") {
      Array.isArray(e[r]) ? e[r].push(o) : (e[r] && r++, (e[r] = [o]));
      continue;
    }
  }
  return ds(e, t);
}
function ds(n, t = "chain") {
  for (const e of n)
    if (Array.isArray(e)) {
      if (e.length === 1) continue;
      let r = 0;
      for (; r < e.length && e[r].type === "space"; ) r++;
      let s = e.length - 1;
      for (; s >= 0 && e[s].type === "space"; ) s--;
      if (r <= s) {
        if (r === s) continue;
        ((e[r].group = "start"), (e[s].group = "end"));
        for (let i = r + 1; i < s; i++)
          e[i].type !== "space" && (e[i].group = "middle");
      }
    }
  return hs(n, t);
}
function hs(n, t = "chain") {
  if (t === "block") return n.flat();
  {
    const e = [];
    for (const r of n)
      Array.isArray(r)
        ? e.push({ type: "thinking_chain_block", tokens: r, raw: "" })
        : e.push(r);
    return e;
  }
}
function di(n) {
  try {
    return URL.canParse ? URL.canParse(n) : (new URL(n), !0);
  } catch {
    return !1;
  }
}
function hi(n) {
  if (!n || typeof n != "string") return "";
  const t = n.split(`
`);
  let e = "";
  for (const s of t) {
    const i = s.trim();
    if (i && i !== ">") {
      e = i;
      break;
    }
  }
  if (!e) return "";
  let r = e;
  return (
    (r = r.replace(/^#{1,6}\s+/, "")),
    (r = r.replace(/^>\s*/, "")),
    (r = r.replace(/^[\*\-\+]\s+/, "")),
    (r = r.replace(/^\d+\.\s+/, "")),
    (r = r.replace(/^```.*$/, "")),
    (r = r.replace(/^`(.+)`$/, "$1")),
    (r = r.replace(/\*\*(.*?)\*\*/g, "$1")),
    (r = r.replace(/__(.*?)__/g, "$1")),
    (r = r.replace(/\*(.*?)\*/g, "$1")),
    (r = r.replace(/_(.*?)_/g, "$1")),
    (r = r.replace(/~~(.*?)~~/g, "$1")),
    (r = r.replace(/`([^`]+)`/g, "$1")),
    (r = r.replace(/\[([^\]]+)\]\([^\)]+\)/g, "$1")),
    (r = r.replace(/\[([^\]]+)\]\[[^\]]*\]/g, "$1")),
    (r = r.replace(/!\[([^\]]*)\]\([^\)]+\)/g, "$1")),
    (r = r.replace(/<[^>]+>/g, "")),
    (r = r.replace(/\\(.)/g, "$1")),
    r.trim()
  );
}
const pi = (n, t) => {
    let e = 0;
    return function (...r) {
      const s = Date.now();
      if (s - e >= t) return ((e = s), n.apply(this, r));
    };
  },
  gi =
    typeof window < "u" &&
    window.localStorage.getItem("z_ai_session_t") ===
      "322f2d5c-46a6-49ed-991c-d7fa3f42a2ee",
  mi = () =>
    (window.screen.orientation || {}).type ||
    screen.mozOrientation ||
    screen.msOrientation,
  xi = (n) => {
    const t = n.style.cssText;
    return (
      (n.style.position = "fixed"),
      (n.style.top = "0"),
      (n.style.left = "0"),
      (n.style.width = "100vw"),
      (n.style.height = "100vh"),
      (n.style.zIndex = "9999"),
      (n.style.backgroundColor = "black"),
      () => {
        n.style.cssText = t;
      }
    );
  };
function bi(n) {
  const t = new Set();
  return (
    t.add("/"),
    n.forEach((e) => {
      if (!e.startsWith("app/") && !e.startsWith("src/app/")) return;
      const r = e.replace(/^(src\/)?app\//, "");
      if (!ps(r)) return;
      const s = gs(r);
      s && t.add(s);
    }),
    Array.from(t).sort()
  );
}
function ps(n) {
  const t = n.split("/").pop() || "";
  return t.startsWith("route.") ||
    [
      "layout.",
      "loading.",
      "error.",
      "not-found.",
      "global-error.",
      "template.",
      "default.",
    ].some((r) => t.startsWith(r))
    ? !1
    : /^page\.(tsx?|jsx?)$/.test(t);
}
function gs(n) {
  const t = n.split("/").slice(0, -1);
  if (t.length === 0) return "/";
  const e = [];
  for (const s of t)
    if (s)
      if (s.startsWith("[") && s.endsWith("]")) {
        const i = s.slice(1, -1);
        i.startsWith("...")
          ? e.push(`[...${i.slice(3)}]`)
          : i.startsWith("[...") && i.endsWith("]")
            ? e.push(`[[...${i.slice(4, -1)}]]`)
            : e.push(`[${i}]`);
      } else {
        if (s.startsWith("(") && s.endsWith(")")) continue;
        e.push(s);
      }
  const r = "/" + e.join("/");
  return r === "/" ? "/" : r;
}
function vi(n, t, e) {
  const r = n;
  let s = 0,
    i = 0;
  for (const o of r) {
    const l = o.toA - o.fromA,
      a = o.toB - o.fromB;
    if (l > 0) {
      const c = e.slice(o.fromA, o.toA),
        m = (c.match(/\n/g) || []).length;
      i += m > 0 ? m : c.length > 0 ? 1 : 0;
    }
    if (a > 0) {
      const c = t.slice(o.fromB, o.toB),
        m = (c.match(/\n/g) || []).length;
      s += m > 0 ? m : c.length > 0 ? 1 : 0;
    }
  }
  return { addedLines: s, deletedLines: i };
}
function nt(n, t) {
  const e = requestIdleCallback(n, t);
  return () => {
    cancelIdleCallback(e);
  };
}
class yi {
  constructor(t, e) {
    ee(this, "message");
    ee(this, "handler");
    ee(this, "toolCallStreamParser", null);
    ee(this, "chatCompletionsQueue", []);
    ee(this, "handleCompletionsQueue", (t) => {
      var e, r, s;
      if (
        !(
          this.chatCompletionsQueue.length === 0 &&
          (e = this.message) != null &&
          e.done
        )
      ) {
        for (
          ;
          t.timeRemaining() > 4 && this.chatCompletionsQueue.length > 0;
        ) {
          const i = this.chatCompletionsQueue.shift();
          if (i && i.delta_content) {
            let o = i.delta_content;
            const l = i.phase;
            for (
              ;
              this.chatCompletionsQueue.length > 0 &&
              this.chatCompletionsQueue[0] &&
              typeof this.chatCompletionsQueue[0].delta_content == "string" &&
              ((r = this.chatCompletionsQueue[0]) == null
                ? void 0
                : r.phase) === l &&
              ((s = this.chatCompletionsQueue[0]) == null
                ? void 0
                : s.phase) !== "tool_response";
            ) {
              const a = this.chatCompletionsQueue.shift();
              o += a.delta_content;
            }
            i.delta_content = o;
          } else if (i && i.content) {
            let o = i;
            for (
              ;
              this.chatCompletionsQueue.length > 0 &&
              this.chatCompletionsQueue[0] &&
              this.chatCompletionsQueue[0].content;
            )
              o = this.chatCompletionsQueue.shift();
            i.content = o.content;
          }
          this.handler(i, this);
        }
        nt(this.handleCompletionsQueue, { timeout: 2e3 });
      }
    });
    ((this.message = t), (this.handler = e));
  }
  run() {
    nt(this.handleCompletionsQueue, { timeout: 2e3 });
  }
}
const ms = (n, t) => {
    let e;
    return (...r) => {
      (clearTimeout(e),
        (e = setTimeout(() => {
          n(...r);
        }, t)));
    };
  },
  _i = (n) => {
    try {
      return new URL(n).pathname.split("/").pop() ?? "";
    } catch {
      return "";
    }
  },
  wi = (n) => {
    try {
      return xs(n);
    } catch {
      return "";
    }
  };
function xs(n) {
  try {
    let t = new URL(n).hostname;
    t.startsWith("www.") && (t = t.slice(4));
    const e = t.split("."),
      r = ["com.cn", "net.cn", "org.cn", "gov.cn", "edu.cn", "co.uk"],
      s = e.slice(-2).join(".");
    let i;
    return (
      r.includes(s) ? (i = e.length - 3) : (i = e.length - 2),
      i <= 0 ? e[0] : e.slice(0, i + 1).join(".")
    );
  } catch {
    return "";
  }
}
function st(n) {
  return new RegExp("\\p{Script=Han}", "u").test(n);
}
function bs(n) {
  return (
    (n = n
      .split(
        `
`,
      )
      .map(
        (r) => (
          /[\u4e00-\u9fa5]/.test(r) &&
            r.includes("*") &&
            (/（|）/.test(r) &&
              ((r = X(r, "**", "（", "）")), (r = X(r, "*", "（", "）"))),
            /“|”/.test(r) &&
              ((r = X(r, "**", "“", "”")), (r = X(r, "*", "“", "”"))),
            /：/.test(r) &&
              ((r = X(r, "**", "：", "：")), (r = X(r, "*", "：", "："))),
            /《|》/.test(r) &&
              ((r = X(r, "**", "《", "》")), (r = X(r, "*", "《", "》")))),
          r
        ),
      ).join(`
`)),
    n
  );
}
function X(n, t, e, r) {
  const s = ns(t),
    i = new RegExp(`(.?)(?<!${s})(${s})([^${s}]+)(${s})(?!${s})(.)`, "g");
  return n.replace(i, (o, l, a, c, m, v) =>
    (c.startsWith(e) && l && l.length > 0 && st(l[l.length - 1])) ||
    (c.endsWith(r) && v && v.length > 0 && st(v[0]))
      ? `${l} ${a}${c}${m} ${v}`
      : o,
  );
}
const it = (n, t = null) => {
    try {
      return JSON.parse(n);
    } catch {
      return t;
    }
  },
  Si = (n) =>
    n
      ? [
          {
            type: "ppt_preset_template",
            file: {
              id: n.id,
              user_id: n.id,
              hash: null,
              filename:
                n.i18n[ne(Ie).language === "zh-CN" ? "cn" : "en"].title +
                ".pptx",
              data: {},
              meta: {
                name:
                  n.i18n[ne(Ie).language === "zh-CN" ? "cn" : "en"].title +
                  ".pptx",
                content_type:
                  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
                size: 28288,
                data: {},
                oss_endpoint: "bj",
                cdn_url: n.images[0],
              },
              created_at: 1145141919810,
              updated_at: 1145141919810,
            },
            id: n.id,
            url: n.images[0],
            name:
              n.i18n[ne(Ie).language === "zh-CN" ? "cn" : "en"].title + ".pptx",
            status: "uploaded",
            size: 28288,
            error: "",
            itemId: n.id,
            media: "doc",
          },
        ]
      : [];
function Ei(n) {
  n.forEach((t) => {
    t.type === "tool_calls" &&
      (t.content.forEach((e) => {
        e.function.parsed_arguments = it(e.function.arguments, {});
      }),
      t.results.forEach((e) => {
        e.content_items = it(e.content, []);
      }));
  });
}
const Ae = () => {
    try {
      const t = new URLSearchParams(window.location.search).get("utm_id");
      if (t && t.trim()) {
        const r = t.trim();
        return (localStorage.setItem("utm_id", r), r);
      }
      const e = localStorage.getItem("utm_id");
      return e || "default";
    } catch (n) {
      return (console.error("[Analytics] Error:", n), "default");
    }
  },
  ot = (n) => {
    const t = Ae();
    return n.includes("?") ? `${n}&fr=${t}` : `${n}?fr=${t}`;
  },
  vs = () => Z.isMobile || Z.isTablet,
  at = (n, t, e = !1) => {
    var s;
    if (t.includes("google-analytics")) return;
    const r = {
      bt: "er",
      md: "network",
      ct: e ? "global_network_error" : "network_error",
      ctvl: n,
      request_url: t,
      usid: ((s = ne(Fe)) == null ? void 0 : s.id) || "anonymous",
      fr: Ae(),
    };
    window.setTimeout(() => {
      ie(r);
    }, 1e3);
  },
  ie = async (n) => {
    try {
      const t = ne(Fe),
        e = n.usid || (t == null ? void 0 : t.id) || "anonymous",
        r = (t == null ? void 0 : t.role) == "guest" ? "1" : "0",
        s = Ae(),
        i = "https://analysis.chatglm.cn/bdms/p.gif",
        o = new URLSearchParams({
          pd: "zai",
          bt: n.bt,
          tm: Z.isAndroid ? "android" : Z.isIOS ? "ios" : "pc",
          ct: n.ct,
          ctvl: n.ctvl ?? "",
          usid: e,
          fr: s,
          pvid: n.pvid || "",
          _n_is_guest: r,
          url: ot(n.url || window.location.pathname),
          ctnm: n.ctnm ?? "",
          extra: n.extra ?? "",
          pds: n.pds ?? "",
          pdt: n.pdt ?? "",
          ctid: n.ctid ?? "",
          ...n.data,
        }),
        l = new Image();
      ((l.src = `${i}?${o.toString()}`),
        n.bt === "cl" &&
          window.gtag("event", n.ct, {
            module: n.md,
            page_type: vs() ? "h5" : "pc",
            value: n.ctvl ?? "",
            pvid: n.pvid || "",
            usid: e,
            fr: s,
            url: ot(n.url || window.location.pathname),
            is_guest: r,
            ctnm: n.ctnm ?? "",
            extra: n.extra ?? "",
            pds: n.pds ?? "",
            pdt: n.pdt ?? "",
            ctid: n.ctid ?? "",
          }));
    } catch (t) {
      console.error("[Analytics] Error:", t);
    }
  },
  ke = (n, t, e, r, s, i, o, l) => {
    ie({
      bt: "pv",
      md: n,
      ct: t,
      ctvl: e,
      ctnm: r ?? "",
      extra: s ?? "",
      pds: i ?? "",
      pdt: o ?? "",
      ctid: l ?? "",
    });
  },
  Oi = (n, t, e, r, s, i, o, l) => {
    ie({
      bt: "cl",
      md: n,
      ct: t,
      ctvl: e,
      ctnm: r ?? "",
      extra: s ?? "",
      pds: i ?? "",
      pdt: o ?? "",
      ctid: l ?? "",
    });
  },
  Ti = (n, t, e, r, s, i, o, l) => {
    ie({
      bt: "RE",
      md: n,
      ct: t,
      ctvl: e,
      ctnm: r,
      extra: s,
      pds: i,
      pdt: o ?? "",
      ctid: l ?? "",
    });
  },
  Li = (n, t, e, r, s, i, o, l) => {
    ie({
      bt: "pf",
      md: n,
      ct: t,
      ctvl: e,
      ctnm: "",
      extra: "",
      pds: "",
      pdt: "",
      ctid: "",
    });
  },
  ys = (n, t, e, r, s, i, o, l) => {
    ie({
      bt: "cl",
      md: n,
      ct: t,
      ctvl: e,
      ctnm: r ?? "",
      extra: s ?? "",
      pds: i ?? "",
      pdt: o ?? "",
      ctid: l ?? "",
    });
  },
  Ci = ms(ys, 2e3);
let Ee,
  Oe = !1;
const ki = () => {
    if (!Oe)
      try {
        ((Ee = window.fetch),
          (window.fetch = async (n, t) => {
            const e =
              typeof n == "string" ? n : n instanceof URL ? n.href : n.url;
            try {
              const r = await Ee(n, t);
              if (!r.ok) {
                const s = `HTTP status Error: ${r.status} ${r.statusText}`;
                at(s, e, !0);
              }
              return r;
            } catch (r) {
              const s = r instanceof Error ? r.message : "Unknown fetch error";
              throw (at(s, e, !0), r);
            }
          }),
          (Oe = !0));
      } catch (n) {
        console.error(
          "[Analytics] Failed to initialize global fetch error tracking:",
          n,
        );
      }
  },
  Ri = () => {
    if (Oe)
      try {
        (Ee && (window.fetch = Ee),
          (Oe = !1),
          console.log("[Analytics] Global fetch error tracking destroyed"));
      } catch (n) {
        console.error(
          "[Analytics] Failed to destroy global fetch error tracking:",
          n,
        );
      }
  };
var _s = ((n) => (
    (n.NewChat = "new_chat"),
    (n.History = "history"),
    (n.FileUpload = "file_upload"),
    (n.Share = "share"),
    (n.Vision = "vision"),
    n
  ))(_s || {}),
  ws = ((n) => ((n.Home = "home"), (n.Chat = "chat"), n))(ws || {});
const Ii = k(qr),
  Pi = k(void 0),
  Fe = k(void 0),
  Ai = k(!1),
  Fi = k({}),
  Ni = k(!1),
  Di = k(null),
  Wi = k(null),
  Mi = k(null),
  Vi = k("system"),
  wt = k("en-US"),
  ji = k("light"),
  Ui = Qr(Gr()),
  $i = k(""),
  Bi = k(""),
  zi = k(""),
  Hi = k(!1),
  Yi = k(!1),
  Ki = k(""),
  qi = k([]),
  Qi = k(!1),
  Gi = k("simple"),
  Ji = k(null),
  Xi = k([]),
  Zi = k([]),
  eo = k([]),
  to = k(null),
  ro = k(null),
  no = k(null),
  so = k([]),
  io = k([]),
  oo = k({ chatDirection: "auto" }),
  ao = k(null),
  lo = k(!1),
  uo = k(!1),
  co = k(!1),
  fo = k(!1),
  ho = k(!1),
  Ss = k(!1),
  po = k(!1);
Ss.subscribe((n) => {
  n && ke("Login guide", "Function login guidance_Exposure", "");
});
const go = k(["new_chat", "home"]),
  mo = k(!1),
  Es = k(!1);
Es.subscribe((n) => {
  n && ke("Login guide", "Scene 1 login guidance_Exposure", "");
});
const Os = k(!1);
Os.subscribe((n) => {
  n && ke("Login guide", "Scene 2 workspace guidance_Exposure", "");
});
const Ts = k(!1);
Ts.subscribe((n) => {
  n && ke("Update pop-up", "Update pop-up_Exposure", "");
});
const xo = k(""),
  bo = k(!1),
  vo = k(!1),
  yo = k(!0),
  _o = k(!1),
  wo = k(!0),
  So = k(!1),
  Eo = k(!1),
  Oo = k(1),
  To = k(!0),
  Lo = k(null),
  Co = k("prod-fe-1.0.252".includes("staging") ? "staging" : "prod");
export {
  Lo as $,
  $i as A,
  xo as B,
  Qi as C,
  cn as D,
  ao as E,
  cs as F,
  Ks as G,
  Bi as H,
  qi as I,
  Gi as J,
  Bs as K,
  zs as L,
  ii as M,
  Vi as N,
  mi as O,
  Z as P,
  xi as Q,
  Zr as R,
  so as S,
  os as T,
  Fs as U,
  Oo as V,
  Ii as W,
  Ji as X,
  Eo as Y,
  Gs as Z,
  Xs as _,
  oo as a,
  ri as a$,
  As as a0,
  To as a1,
  ki as a2,
  Ai as a3,
  Ie as a4,
  Ri as a5,
  zi as a6,
  Xi as a7,
  co as a8,
  So as a9,
  ie as aA,
  Li as aB,
  Qs as aC,
  wi as aD,
  Us as aE,
  wo as aF,
  di as aG,
  gi as aH,
  hi as aI,
  fi as aJ,
  Ms as aK,
  js as aL,
  ei as aM,
  Zs as aN,
  Vs as aO,
  li as aP,
  Ys as aQ,
  _i as aR,
  Wi as aS,
  Mi as aT,
  Ci as aU,
  pi as aV,
  yt as aW,
  bi as aX,
  Di as aY,
  yo as aZ,
  Yi as a_,
  Hi as aa,
  Ki as ab,
  Ss as ac,
  go as ad,
  _s as ae,
  _o as af,
  bo as ag,
  vo as ah,
  ws as ai,
  uo as aj,
  oi as ak,
  ai as al,
  fo as am,
  io as an,
  to as ao,
  ro as ap,
  Fi as aq,
  $s as ar,
  Es as as,
  Os as at,
  Ws as au,
  Ti as av,
  it as aw,
  po as ax,
  Co as ay,
  Si as az,
  qs as b,
  Ns as b0,
  Ds as b1,
  ti as b2,
  ci as b3,
  yi as b4,
  Pi as c,
  at as d,
  si as e,
  no as f,
  ho as g,
  ni as h,
  Js as i,
  mo as j,
  ji as k,
  wt as l,
  eo as m,
  Ni as n,
  Oi as o,
  Hs as p,
  Ps as q,
  ui as r,
  lo as s,
  ke as t,
  Fe as u,
  ut as v,
  Ei as w,
  vi as x,
  Zi as y,
  Ui as z,
};
