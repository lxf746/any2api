import { b as s, O as A } from "./BJ4IMYPr.js";
import { l, g as S, d as C, q as b, r as L } from "./CAm9rDEa.js";
import { U as r } from "./DFZQlWS9.js";
import "./BzqETyZD.js";
import { t as E } from "./BrXpCOvA.js";
const v = async (e = "") => {
    let o = null;
    const a = await fetch(`${A}/config`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        ...(e && { authorization: `Bearer ${e}` }),
      },
    })
      .then(async (t) => {
        if (!t.ok) throw await t.json();
        return t.json();
      })
      .catch(
        (t) => (
          console.log(t),
          "detail" in t ? (o = t.detail) : (o = "Server connection failed"),
          null
        ),
      );
    if (o) throw o;
    return a;
  },
  J = async (e = "", o) => {
    let a = null;
    const t = await fetch(`${A}/config/update`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        ...(e && { authorization: `Bearer ${e}` }),
      },
      body: JSON.stringify({ ...o }),
    })
      .then(async (n) => {
        if (!n.ok) throw await n.json();
        return n.json();
      })
      .catch(
        (n) => (
          console.log(n),
          "detail" in n ? (a = n.detail) : (a = "Server connection failed"),
          null
        ),
      );
    if (a) throw a;
    return t;
  },
  k = async (e, o) => {
    let a = null;
    const t = await fetch(`${e}/models`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        ...(o && { authorization: `Bearer ${o}` }),
      },
    })
      .then(async (n) => {
        if (!n.ok) throw await n.json();
        return n.json();
      })
      .catch((n) => {
        var i;
        return (
          (a = `OpenAI: ${((i = n == null ? void 0 : n.error) == null ? void 0 : i.message) ?? "Network Problem"}`),
          []
        );
      });
    if (a) throw a;
    return t;
  },
  R = async (e, o) => {
    let a = null;
    const t = await fetch(`${A}/models${typeof o == "number" ? `/${o}` : ""}`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        ...(e && { authorization: `Bearer ${e}` }),
      },
    })
      .then(async (n) => {
        if (!n.ok) throw await n.json();
        return n.json();
      })
      .catch((n) => {
        var i;
        return (
          (a = `OpenAI: ${((i = n == null ? void 0 : n.error) == null ? void 0 : i.message) ?? "Network Problem"}`),
          []
        );
      });
    if (a) throw a;
    return t;
  },
  x = async (e = "", o = "https://api.openai.com/v1", a = "", t = !1) => {
    if (!o) throw "OpenAI: URL is required";
    let n = null,
      i = null;
    if (t) {
      if (
        ((i = await fetch(`${o}/models`, {
          method: "GET",
          headers: {
            Accept: "application/json",
            Authorization: `Bearer ${a}`,
            "Content-Type": "application/json",
          },
        })
          .then(async (c) => {
            if (!c.ok) throw await c.json();
            return c.json();
          })
          .catch((c) => {
            var h;
            return (
              (n = `OpenAI: ${((h = c == null ? void 0 : c.error) == null ? void 0 : h.message) ?? "Network Problem"}`),
              []
            );
          })),
        n)
      )
        throw n;
    } else if (
      ((i = await fetch(`${A}/verify`, {
        method: "POST",
        headers: {
          Accept: "application/json",
          Authorization: `Bearer ${e}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ url: o, key: a }),
      })
        .then(async (c) => {
          if (!c.ok) throw await c.json();
          return c.json();
        })
        .catch((c) => {
          var h;
          return (
            (n = `OpenAI: ${((h = c == null ? void 0 : c.error) == null ? void 0 : h.message) ?? "Network Problem"}`),
            []
          );
        })),
      n)
    )
      throw n;
    return i;
  },
  V = async (e = "", o, a = `${s}/api`, t) => {
    let n = null;
    const i = await fetch(`${a}/chat/continue`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${e}`,
        "Content-Type": "application/json",
        "Accept-Language": r(l) ?? "en-US",
        "X-FE-Version": "prod-fe-1.0.252",
      },
      body: JSON.stringify({ message_id: o }),
      signal: t.signal,
    })
      .then(async (c) => {
        if (
          (c.status === 401 && S.set(!0),
          c.status === 426 &&
            E.error(
              b(
                "New version detected, please refresh the page to get the latest features",
              ),
              {
                action: {
                  label: b("Refresh Now"),
                  onClick: () => window.location.reload(),
                },
                duration: 1 / 0,
              },
            ),
          !c.ok)
        )
          throw await c.json();
        if (!c.body) throw new Error("No response body");
        return c;
      })
      .catch((c) => ((n = `${(c == null ? void 0 : c.detail) ?? c}`), null));
    if (n) {
      const c = `${(n == null ? void 0 : n.detail) ?? n}`;
      throw (C(c, `${a}/chat/continue`), n);
    }
    return i;
  },
  D = async (e = "", o, a = `${s}/api`, t, n = "", i = "") => {
    let c = null;
    const h = await fetch(`${a}/chat/completions?${i}`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${e}`,
        "Content-Type": "application/json",
        "Accept-Language": r(l) ?? "en-US",
        "X-FE-Version": "prod-fe-1.0.252",
        "X-Signature": n,
      },
      body: JSON.stringify(o),
      signal: t.signal,
    })
      .then(async (p) => {
        if ((p.status === 401 && S.set(!0), !p.ok)) throw await p.json();
        if (!p.body) throw new Error("No response body");
        return p;
      })
      .catch((p) => ((c = `${(p == null ? void 0 : p.detail) ?? p}`), null));
    if (c) {
      const p = `${(c == null ? void 0 : c.detail) ?? c}`;
      throw (C(p, `${a}/chat/completions`), c);
    }
    return h;
  },
  F = async (e = "", o = null, a = !1, t) => {
    const n = e || localStorage.getItem("token") || "";
    let i = null;
    const c = await (
      t ||
      fetch(`${s}/api/models${a ? "/base" : ""}`, {
        method: "GET",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
          "Accept-Language": r(l),
          ...(n && { authorization: `Bearer ${n}` }),
        },
      })
    )
      .then(async (p) => {
        if ((p.status === 401 && S.set(!0), !p.ok)) throw await p.json();
        return p.json();
      })
      .catch((p) => ((i = p), console.log(p), null));
    if (i) throw i;
    let h = (c == null ? void 0 : c.data) ?? [];
    if (o && !a) {
      let p = [];
      if (o) {
        const u = o.OPENAI_API_BASE_URLS,
          B = o.OPENAI_API_KEYS,
          T = o.OPENAI_API_CONFIGS,
          $ = [];
        for (const d in u) {
          const g = u[d];
          if (d.toString() in T) {
            const w = T[d.toString()] ?? {},
              y = (w == null ? void 0 : w.enable) ?? !0,
              m = (w == null ? void 0 : w.model_ids) ?? [];
            if (y)
              if (m.length > 0) {
                const j = {
                  object: "list",
                  data: m.map((f) => ({
                    id: f,
                    name: f,
                    owned_by: "openai",
                    openai: { id: f },
                    urlIdx: d,
                  })),
                };
                $.push((async () => j)());
              } else
                $.push(
                  (async () =>
                    await k(g, B[d])
                      .then((j) => j)
                      .catch((j) => ({
                        object: "list",
                        data: [],
                        urlIdx: d,
                      })))(),
                );
            else
              $.push((async () => ({ object: "list", data: [], urlIdx: d }))());
          }
        }
        const O = await Promise.all($);
        for (const d in O) {
          const g = O[d],
            w = T[d.toString()] ?? {};
          let y = Array.isArray(g) ? g : ((g == null ? void 0 : g.data) ?? []);
          y = y.map((f) => ({ ...f, openai: { id: f.id }, urlIdx: d }));
          const m = w.prefix_id;
          if (m) for (const f of y) f.id = `${m}.${f.id}`;
          const j = w.tags;
          if (j) for (const f of y) f.tags = j;
          p = p.concat(y);
        }
      }
      h = h.concat(
        p.map((u) => ({
          ...u,
          name: (u == null ? void 0 : u.name) ?? (u == null ? void 0 : u.id),
          direct: !0,
        })),
      );
      const P = {};
      for (const u of h) P[u.id] = u;
      h = Object.values(P);
    }
    return h;
  },
  M = async (e, o, a) => {
    let t = null;
    const n = await fetch(`${s}/api/chat/actions/${o}`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        ...(e && { authorization: `Bearer ${e}` }),
      },
      body: JSON.stringify(a),
    })
      .then(async (i) => {
        if (!i.ok) throw await i.json();
        return i.json();
      })
      .catch(
        (i) => (console.log(i), "detail" in i ? (t = i.detail) : (t = i), null),
      );
    if (t) throw t;
    return n;
  },
  q = async (e, o, a) => {
    let t = null;
    const n = await fetch(`${s}/api/tasks/stop/${o}`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        ...(e && { authorization: `Bearer ${e}` }),
      },
      body: JSON.stringify({ reason: a }),
    })
      .then(async (i) => {
        if (!i.ok) throw await i.json();
        return i.json();
      })
      .catch(
        (i) => (console.log(i), "detail" in i ? (t = i.detail) : (t = i), null),
      );
    if (t) throw t;
    return n;
  },
  z = async (e, o) => {
    let a = null;
    const t = await fetch(`${o}`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        ...(e && { authorization: `Bearer ${e}` }),
      },
    })
      .then(async (i) => {
        if (!i.ok) throw await i.json();
        return i.json();
      })
      .catch(
        (i) => (console.log(i), "detail" in i ? (a = i.detail) : (a = i), null),
      );
    if (a) throw a;
    const n = { openapi: t, info: t.info, specs: L(t) };
    return (console.log(n), n);
  },
  W = async (e, o) =>
    (
      await Promise.all(
        o
          .filter((a) => {
            var t;
            return (t = a == null ? void 0 : a.config) == null
              ? void 0
              : t.enable;
          })
          .map(async (a) => {
            const t = await z(
              a == null ? void 0 : a.key,
              (a == null ? void 0 : a.url) +
                "/" +
                ((a == null ? void 0 : a.path) ?? "openapi.json"),
            ).catch(
              (n) => (
                E.error(
                  e.t("Failed to connect to {{URL}} OpenAPI tool server", {
                    URL:
                      (a == null ? void 0 : a.url) +
                      "/" +
                      ((a == null ? void 0 : a.path) ?? "openapi.json"),
                  }),
                ),
                null
              ),
            );
            if (t) {
              const { openapi: n, info: i, specs: c } = t;
              return {
                url: a == null ? void 0 : a.url,
                openapi: n,
                info: i,
                specs: c,
              };
            }
          }),
      )
    ).filter((a) => a),
  X = async (e = "") => {
    let o = null;
    const a = await fetch(`${s}/api/v1/tasks/config`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        ...(e && { authorization: `Bearer ${e}` }),
      },
    })
      .then(async (t) => {
        if (!t.ok) throw await t.json();
        return t.json();
      })
      .catch((t) => (console.log(t), (o = t), null));
    if (o) throw o;
    return a;
  },
  K = async (e, o) => {
    let a = null;
    const t = await fetch(`${s}/api/v1/tasks/config/update`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        ...(e && { authorization: `Bearer ${e}` }),
      },
      body: JSON.stringify(o),
    })
      .then(async (n) => {
        if (!n.ok) throw await n.json();
        return n.json();
      })
      .catch(
        (n) => (console.log(n), "detail" in n ? (a = n.detail) : (a = n), null),
      );
    if (a) throw a;
    return t;
  },
  Y = async (e = "") => {
    let o = null;
    const a = await fetch(`${s}/api/v1/pipelines/list`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        ...(e && { authorization: `Bearer ${e}` }),
      },
    })
      .then(async (n) => {
        if (!n.ok) throw await n.json();
        return n.json();
      })
      .catch((n) => (console.log(n), (o = n), null));
    if (o) throw o;
    return (a == null ? void 0 : a.data) ?? [];
  },
  H = async (e, o, a) => {
    let t = null;
    const n = new FormData();
    (n.append("file", o), n.append("urlIdx", a));
    const i = await fetch(`${s}/api/v1/pipelines/upload`, {
      method: "POST",
      headers: {
        "Accept-Language": r(l),
        ...(e && { authorization: `Bearer ${e}` }),
      },
      body: n,
    })
      .then(async (c) => {
        if (!c.ok) throw await c.json();
        return c.json();
      })
      .catch(
        (c) => (console.log(c), "detail" in c ? (t = c.detail) : (t = c), null),
      );
    if (t) throw t;
    return i;
  },
  Q = async (e, o, a) => {
    let t = null;
    const n = await fetch(`${s}/api/v1/pipelines/add`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        ...(e && { authorization: `Bearer ${e}` }),
      },
      body: JSON.stringify({ url: o, urlIdx: a }),
    })
      .then(async (i) => {
        if (!i.ok) throw await i.json();
        return i.json();
      })
      .catch(
        (i) => (console.log(i), "detail" in i ? (t = i.detail) : (t = i), null),
      );
    if (t) throw t;
    return n;
  },
  Z = async (e, o, a) => {
    let t = null;
    const n = await fetch(`${s}/api/v1/pipelines/delete`, {
      method: "DELETE",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        ...(e && { authorization: `Bearer ${e}` }),
      },
      body: JSON.stringify({ id: o, urlIdx: a }),
    })
      .then(async (i) => {
        if (!i.ok) throw await i.json();
        return i.json();
      })
      .catch(
        (i) => (console.log(i), "detail" in i ? (t = i.detail) : (t = i), null),
      );
    if (t) throw t;
    return n;
  },
  tt = async (e, o) => {
    let a = null;
    const t = new URLSearchParams();
    o !== void 0 && t.append("urlIdx", o);
    const n = await fetch(`${s}/api/v1/pipelines/?${t.toString()}`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        ...(e && { authorization: `Bearer ${e}` }),
      },
    })
      .then(async (c) => {
        if (!c.ok) throw await c.json();
        return c.json();
      })
      .catch((c) => (console.log(c), (a = c), null));
    if (a) throw a;
    return (n == null ? void 0 : n.data) ?? [];
  },
  nt = async (e, o, a) => {
    let t = null;
    const n = new URLSearchParams();
    a !== void 0 && n.append("urlIdx", a);
    const i = await fetch(`${s}/api/v1/pipelines/${o}/valves?${n.toString()}`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        ...(e && { authorization: `Bearer ${e}` }),
      },
    })
      .then(async (c) => {
        if (!c.ok) throw await c.json();
        return c.json();
      })
      .catch((c) => (console.log(c), (t = c), null));
    if (t) throw t;
    return i;
  },
  at = async (e, o, a) => {
    let t = null;
    const n = new URLSearchParams();
    a !== void 0 && n.append("urlIdx", a);
    const i = await fetch(
      `${s}/api/v1/pipelines/${o}/valves/spec?${n.toString()}`,
      {
        method: "GET",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
          "Accept-Language": r(l),
          ...(e && { authorization: `Bearer ${e}` }),
        },
      },
    )
      .then(async (c) => {
        if (!c.ok) throw await c.json();
        return c.json();
      })
      .catch((c) => (console.log(c), (t = c), null));
    if (t) throw t;
    return i;
  },
  ot = async (e = "", o, a, t) => {
    let n = null;
    const i = new URLSearchParams();
    t !== void 0 && i.append("urlIdx", t);
    const c = await fetch(
      `${s}/api/v1/pipelines/${o}/valves/update?${i.toString()}`,
      {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Accept-Language": r(l),
          "Content-Type": "application/json",
          ...(e && { authorization: `Bearer ${e}` }),
        },
        body: JSON.stringify(a),
      },
    )
      .then(async (h) => {
        if (!h.ok) throw await h.json();
        return h.json();
      })
      .catch(
        (h) => (console.log(h), "detail" in h ? (n = h.detail) : (n = h), null),
      );
    if (n) throw n;
    return c;
  },
  et = async () => {
    let e = null;
    const o = localStorage.getItem("token"),
      a = await fetch(`${s}/api/config`, {
        method: "GET",
        credentials: "include",
        headers: {
          "Accept-Language": r(l),
          "Content-Type": "application/json",
          ...(o ? { authorization: `Bearer ${o}` } : {}),
        },
      })
        .then(async (t) => {
          if (!t.ok) throw await t.json();
          return t.json();
        })
        .catch((t) => (console.log(t), (e = t), null));
    if (e) throw e;
    return a;
  },
  it = async (e) => {
    let o = null;
    const a = await fetch(`${s}/api/version/updates`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        Authorization: `Bearer ${e}`,
      },
    })
      .then(async (t) => {
        if (!t.ok) throw await t.json();
        return t.json();
      })
      .catch((t) => (console.log(t), (o = t), null));
    if (o) throw o;
    return a;
  },
  ct = async (e) => {
    let o = null;
    const a = await fetch(`${s}/api/webhook`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${e}`,
      },
    })
      .then(async (t) => {
        if (!t.ok) throw await t.json();
        return t.json();
      })
      .catch((t) => (console.log(t), (o = t), null));
    if (o) throw o;
    return a.url;
  },
  st = async (e, o) => {
    let a = null;
    const t = await fetch(`${s}/api/webhook`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${e}`,
      },
      body: JSON.stringify({ url: o }),
    })
      .then(async (n) => {
        if (!n.ok) throw await n.json();
        return n.json();
      })
      .catch((n) => (console.log(n), (a = n), null));
    if (a) throw a;
    return t.url;
  },
  lt = async (e) => {
    let o = null;
    const a = await fetch(`${s}/api/v1/mcp/config`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        Authorization: `Bearer ${e}`,
      },
    })
      .then(async (t) => {
        if (!t.ok) throw await t.json();
        return t.json();
      })
      .catch((t) => (console.log(t), (o = t), null));
    if (o) throw o;
    return a;
  },
  rt = async (e, o) => {
    let a = null;
    const t = await fetch(`${s}/api/v1/mcp/config`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        Authorization: `Bearer ${e}`,
      },
      body: JSON.stringify({ config_yaml: o }),
    })
      .then(async (n) => {
        if (!n.ok) throw await n.json();
        return n.json();
      })
      .catch((n) => (console.log(n), (a = n), null));
    if (a) throw a;
    return t;
  },
  pt = async (e) => {
    let o = null;
    const a = await fetch(`${s}/api/v1/admin/vibe-templates`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        Authorization: `Bearer ${e}`,
      },
    })
      .then(async (t) => {
        if (!t.ok) throw await t.json();
        return t.json();
      })
      .catch((t) => (console.log(t), (o = t), null));
    if (o) throw o;
    return a;
  },
  ht = async (e, o) => {
    let a = null;
    const t = await fetch(`${s}/api/v1/admin/vibe-templates/`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        Authorization: `Bearer ${e}`,
      },
      body: JSON.stringify(o),
    })
      .then(async (n) => {
        if (!n.ok) throw await n.json();
        return n.json();
      })
      .catch((n) => (console.log(n), (a = n), null));
    if (a) throw a;
    return t;
  },
  ut = async (e, o, a) => {
    let t = null;
    const n = await fetch(`${s}/api/v1/admin/vibe-templates/${o}/update`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        Authorization: `Bearer ${e}`,
      },
      body: JSON.stringify(a),
    })
      .then(async (i) => {
        if (!i.ok) throw await i.json();
        return i.json();
      })
      .catch((i) => (console.log(i), (t = i), null));
    if (t) throw t;
    return n;
  },
  dt = async (e, o) => {
    let a = null;
    const t = await fetch(`${s}/api/v1/admin/vibe-templates/${o}`, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        Authorization: `Bearer ${e}`,
      },
    })
      .then(async (n) => {
        if (!n.ok) throw await n.json();
        return n.json();
      })
      .catch((n) => (console.log(n), (a = n), null));
    if (a) throw a;
    return t;
  },
  ft = async (e) => {
    let o = null;
    const a = await fetch(`${s}/api/v1/admin/vibe-templates/initialize`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": r(l),
        Authorization: `Bearer ${e}`,
      },
    })
      .then(async (t) => {
        if (!t.ok) throw await t.json();
        return t.json();
      })
      .catch((t) => (console.log(t), (o = t), null));
    if (o) throw o;
    return a;
  };
export {
  dt as A,
  ft as B,
  D as C,
  M as D,
  q as E,
  V as F,
  et as a,
  z as b,
  W as c,
  it as d,
  ct as e,
  Y as f,
  F as g,
  ot as h,
  at as i,
  nt as j,
  tt as k,
  Q as l,
  Z as m,
  H as n,
  X as o,
  K as p,
  v as q,
  R as r,
  J as s,
  lt as t,
  st as u,
  x as v,
  rt as w,
  pt as x,
  ht as y,
  ut as z,
};
