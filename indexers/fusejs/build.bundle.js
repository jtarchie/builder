// search/fuse.js
function t(t2) {
  return Array.isArray ? Array.isArray(t2) : "[object Array]" === o(t2);
}
function e(t2) {
  return "string" == typeof t2;
}
function n(t2) {
  return "number" == typeof t2;
}
function s(t2) {
  return true === t2 || false === t2 || function(t3) {
    return i(t3) && null !== t3;
  }(t2) && "[object Boolean]" == o(t2);
}
function i(t2) {
  return "object" == typeof t2;
}
function r(t2) {
  return null != t2;
}
function c(t2) {
  return !t2.trim().length;
}
function o(t2) {
  return null == t2 ? void 0 === t2 ? "[object Undefined]" : "[object Null]" : Object.prototype.toString.call(t2);
}
var h = Object.prototype.hasOwnProperty;
var a = class {
  constructor(t2) {
    this._keys = [], this._keyMap = {};
    let e2 = 0;
    t2.forEach((t3) => {
      let n2 = l(t3);
      e2 += n2.weight, this._keys.push(n2), this._keyMap[n2.id] = n2, e2 += n2.weight;
    }), this._keys.forEach((t3) => {
      t3.weight /= e2;
    });
  }
  get(t2) {
    return this._keyMap[t2];
  }
  keys() {
    return this._keys;
  }
  toJSON() {
    return JSON.stringify(this._keys);
  }
};
function l(n2) {
  let s2 = null, i2 = null, r2 = null, c2 = 1, o2 = null;
  if (e(n2) || t(n2))
    r2 = n2, s2 = u(n2), i2 = d(n2);
  else {
    if (!h.call(n2, "name"))
      throw new Error(((t3) => `Missing ${t3} property in key`)("name"));
    const t2 = n2.name;
    if (r2 = t2, h.call(n2, "weight") && (c2 = n2.weight, c2 <= 0))
      throw new Error(((t3) => `Property 'weight' in key '${t3}' must be a positive integer`)(t2));
    s2 = u(t2), i2 = d(t2), o2 = n2.getFn;
  }
  return { path: s2, id: i2, weight: c2, src: r2, getFn: o2 };
}
function u(e2) {
  return t(e2) ? e2 : e2.split(".");
}
function d(e2) {
  return t(e2) ? e2.join(".") : e2;
}
var g = { isCaseSensitive: false, includeScore: false, keys: [], shouldSort: true, sortFn: (t2, e2) => t2.score === e2.score ? t2.idx < e2.idx ? -1 : 1 : t2.score < e2.score ? -1 : 1, includeMatches: false, findAllMatches: false, minMatchCharLength: 1, location: 0, threshold: 0.6, distance: 100, ...{ useExtendedSearch: false, getFn: function(i2, c2) {
  let o2 = [], h2 = false;
  const a2 = (i3, c3, l2) => {
    if (r(i3))
      if (c3[l2]) {
        const u2 = i3[c3[l2]];
        if (!r(u2))
          return;
        if (l2 === c3.length - 1 && (e(u2) || n(u2) || s(u2)))
          o2.push(function(t2) {
            return null == t2 ? "" : function(t3) {
              if ("string" == typeof t3)
                return t3;
              let e2 = t3 + "";
              return "0" == e2 && 1 / t3 == -1 / 0 ? "-0" : e2;
            }(t2);
          }(u2));
        else if (t(u2)) {
          h2 = true;
          for (let t2 = 0, e2 = u2.length; t2 < e2; t2 += 1)
            a2(u2[t2], c3, l2 + 1);
        } else
          c3.length && a2(u2, c3, l2 + 1);
      } else
        o2.push(i3);
  };
  return a2(i2, e(c2) ? c2.split(".") : c2, 0), h2 ? o2 : o2[0];
}, ignoreLocation: false, ignoreFieldNorm: false, fieldNormWeight: 1 } };
var f = /[^ ]+/g;
var p = class {
  constructor({ getFn: t2 = g.getFn, fieldNormWeight: e2 = g.fieldNormWeight } = {}) {
    this.norm = function(t3 = 1, e3 = 3) {
      const n2 = /* @__PURE__ */ new Map(), s2 = Math.pow(10, e3);
      return { get(e4) {
        const i2 = e4.match(f).length;
        if (n2.has(i2))
          return n2.get(i2);
        const r2 = 1 / Math.pow(i2, 0.5 * t3), c2 = parseFloat(Math.round(r2 * s2) / s2);
        return n2.set(i2, c2), c2;
      }, clear() {
        n2.clear();
      } };
    }(e2, 3), this.getFn = t2, this.isCreated = false, this.setIndexRecords();
  }
  setSources(t2 = []) {
    this.docs = t2;
  }
  setIndexRecords(t2 = []) {
    this.records = t2;
  }
  setKeys(t2 = []) {
    this.keys = t2, this._keysMap = {}, t2.forEach((t3, e2) => {
      this._keysMap[t3.id] = e2;
    });
  }
  create() {
    !this.isCreated && this.docs.length && (this.isCreated = true, e(this.docs[0]) ? this.docs.forEach((t2, e2) => {
      this._addString(t2, e2);
    }) : this.docs.forEach((t2, e2) => {
      this._addObject(t2, e2);
    }), this.norm.clear());
  }
  add(t2) {
    const n2 = this.size();
    e(t2) ? this._addString(t2, n2) : this._addObject(t2, n2);
  }
  removeAt(t2) {
    this.records.splice(t2, 1);
    for (let e2 = t2, n2 = this.size(); e2 < n2; e2 += 1)
      this.records[e2].i -= 1;
  }
  getValueForItemAtKeyId(t2, e2) {
    return t2[this._keysMap[e2]];
  }
  size() {
    return this.records.length;
  }
  _addString(t2, e2) {
    if (!r(t2) || c(t2))
      return;
    let n2 = { v: t2, i: e2, n: this.norm.get(t2) };
    this.records.push(n2);
  }
  _addObject(n2, s2) {
    let i2 = { i: s2, $: {} };
    this.keys.forEach((s3, o2) => {
      let h2 = s3.getFn ? s3.getFn(n2) : this.getFn(n2, s3.path);
      if (r(h2)) {
        if (t(h2)) {
          let n3 = [];
          const s4 = [{ nestedArrIndex: -1, value: h2 }];
          for (; s4.length; ) {
            const { nestedArrIndex: i3, value: o3 } = s4.pop();
            if (r(o3))
              if (e(o3) && !c(o3)) {
                let t2 = { v: o3, i: i3, n: this.norm.get(o3) };
                n3.push(t2);
              } else
                t(o3) && o3.forEach((t2, e2) => {
                  s4.push({ nestedArrIndex: e2, value: t2 });
                });
          }
          i2.$[o2] = n3;
        } else if (e(h2) && !c(h2)) {
          let t2 = { v: h2, n: this.norm.get(h2) };
          i2.$[o2] = t2;
        }
      }
    }), this.records.push(i2);
  }
  toJSON() {
    return { keys: this.keys, records: this.records };
  }
};
function m(t2, e2, { getFn: n2 = g.getFn, fieldNormWeight: s2 = g.fieldNormWeight } = {}) {
  const i2 = new p({ getFn: n2, fieldNormWeight: s2 });
  return i2.setKeys(t2.map(l)), i2.setSources(e2), i2.create(), i2;
}
function M(t2, { errors: e2 = 0, currentLocation: n2 = 0, expectedLocation: s2 = 0, distance: i2 = g.distance, ignoreLocation: r2 = g.ignoreLocation } = {}) {
  const c2 = e2 / t2.length;
  if (r2)
    return c2;
  const o2 = Math.abs(s2 - n2);
  return i2 ? c2 + o2 / i2 : o2 ? 1 : c2;
}
function x(t2, e2, n2, { location: s2 = g.location, distance: i2 = g.distance, threshold: r2 = g.threshold, findAllMatches: c2 = g.findAllMatches, minMatchCharLength: o2 = g.minMatchCharLength, includeMatches: h2 = g.includeMatches, ignoreLocation: a2 = g.ignoreLocation } = {}) {
  if (e2.length > 32)
    throw new Error(`Pattern length exceeds max of ${32}.`);
  const l2 = e2.length, u2 = t2.length, d2 = Math.max(0, Math.min(s2, u2));
  let f2 = r2, p2 = d2;
  const m2 = o2 > 1 || h2, x2 = m2 ? Array(u2) : [];
  let y2;
  for (; (y2 = t2.indexOf(e2, p2)) > -1; ) {
    let t3 = M(e2, { currentLocation: y2, expectedLocation: d2, distance: i2, ignoreLocation: a2 });
    if (f2 = Math.min(t3, f2), p2 = y2 + l2, m2) {
      let t4 = 0;
      for (; t4 < l2; )
        x2[y2 + t4] = 1, t4 += 1;
    }
  }
  p2 = -1;
  let L2 = [], k2 = 1, _2 = l2 + u2;
  const v2 = 1 << l2 - 1;
  for (let s3 = 0; s3 < l2; s3 += 1) {
    let r3 = 0, o3 = _2;
    for (; r3 < o3; ) {
      M(e2, { errors: s3, currentLocation: d2 + o3, expectedLocation: d2, distance: i2, ignoreLocation: a2 }) <= f2 ? r3 = o3 : _2 = o3, o3 = Math.floor((_2 - r3) / 2 + r3);
    }
    _2 = o3;
    let h3 = Math.max(1, d2 - o3 + 1), g2 = c2 ? u2 : Math.min(d2 + o3, u2) + l2, y3 = Array(g2 + 2);
    y3[g2 + 1] = (1 << s3) - 1;
    for (let r4 = g2; r4 >= h3; r4 -= 1) {
      let c3 = r4 - 1, o4 = n2[t2.charAt(c3)];
      if (m2 && (x2[c3] = +!!o4), y3[r4] = (y3[r4 + 1] << 1 | 1) & o4, s3 && (y3[r4] |= (L2[r4 + 1] | L2[r4]) << 1 | 1 | L2[r4 + 1]), y3[r4] & v2 && (k2 = M(e2, { errors: s3, currentLocation: c3, expectedLocation: d2, distance: i2, ignoreLocation: a2 }), k2 <= f2)) {
        if (f2 = k2, p2 = c3, p2 <= d2)
          break;
        h3 = Math.max(1, 2 * d2 - p2);
      }
    }
    if (M(e2, { errors: s3 + 1, currentLocation: d2, expectedLocation: d2, distance: i2, ignoreLocation: a2 }) > f2)
      break;
    L2 = y3;
  }
  const S2 = { isMatch: p2 >= 0, score: Math.max(1e-3, k2) };
  if (m2) {
    const t3 = function(t4 = [], e3 = g.minMatchCharLength) {
      let n3 = [], s3 = -1, i3 = -1, r3 = 0;
      for (let c3 = t4.length; r3 < c3; r3 += 1) {
        let c4 = t4[r3];
        c4 && -1 === s3 ? s3 = r3 : c4 || -1 === s3 || (i3 = r3 - 1, i3 - s3 + 1 >= e3 && n3.push([s3, i3]), s3 = -1);
      }
      return t4[r3 - 1] && r3 - s3 >= e3 && n3.push([s3, r3 - 1]), n3;
    }(x2, o2);
    t3.length ? h2 && (S2.indices = t3) : S2.isMatch = false;
  }
  return S2;
}
function y(t2) {
  let e2 = {};
  for (let n2 = 0, s2 = t2.length; n2 < s2; n2 += 1) {
    const i2 = t2.charAt(n2);
    e2[i2] = (e2[i2] || 0) | 1 << s2 - n2 - 1;
  }
  return e2;
}
var L = class {
  constructor(t2, { location: e2 = g.location, threshold: n2 = g.threshold, distance: s2 = g.distance, includeMatches: i2 = g.includeMatches, findAllMatches: r2 = g.findAllMatches, minMatchCharLength: c2 = g.minMatchCharLength, isCaseSensitive: o2 = g.isCaseSensitive, ignoreLocation: h2 = g.ignoreLocation } = {}) {
    if (this.options = { location: e2, threshold: n2, distance: s2, includeMatches: i2, findAllMatches: r2, minMatchCharLength: c2, isCaseSensitive: o2, ignoreLocation: h2 }, this.pattern = o2 ? t2 : t2.toLowerCase(), this.chunks = [], !this.pattern.length)
      return;
    const a2 = (t3, e3) => {
      this.chunks.push({ pattern: t3, alphabet: y(t3), startIndex: e3 });
    }, l2 = this.pattern.length;
    if (l2 > 32) {
      let t3 = 0;
      const e3 = l2 % 32, n3 = l2 - e3;
      for (; t3 < n3; )
        a2(this.pattern.substr(t3, 32), t3), t3 += 32;
      if (e3) {
        const t4 = l2 - 32;
        a2(this.pattern.substr(t4), t4);
      }
    } else
      a2(this.pattern, 0);
  }
  searchIn(t2) {
    const { isCaseSensitive: e2, includeMatches: n2 } = this.options;
    if (e2 || (t2 = t2.toLowerCase()), this.pattern === t2) {
      let e3 = { isMatch: true, score: 0 };
      return n2 && (e3.indices = [[0, t2.length - 1]]), e3;
    }
    const { location: s2, distance: i2, threshold: r2, findAllMatches: c2, minMatchCharLength: o2, ignoreLocation: h2 } = this.options;
    let a2 = [], l2 = 0, u2 = false;
    this.chunks.forEach(({ pattern: e3, alphabet: d3, startIndex: g2 }) => {
      const { isMatch: f2, score: p2, indices: m2 } = x(t2, e3, d3, { location: s2 + g2, distance: i2, threshold: r2, findAllMatches: c2, minMatchCharLength: o2, includeMatches: n2, ignoreLocation: h2 });
      f2 && (u2 = true), l2 += p2, f2 && m2 && (a2 = [...a2, ...m2]);
    });
    let d2 = { isMatch: u2, score: u2 ? l2 / this.chunks.length : 1 };
    return u2 && n2 && (d2.indices = a2), d2;
  }
};
var k = class {
  constructor(t2) {
    this.pattern = t2;
  }
  static isMultiMatch(t2) {
    return _(t2, this.multiRegex);
  }
  static isSingleMatch(t2) {
    return _(t2, this.singleRegex);
  }
  search() {
  }
};
function _(t2, e2) {
  const n2 = t2.match(e2);
  return n2 ? n2[1] : null;
}
var v = class extends k {
  constructor(t2, { location: e2 = g.location, threshold: n2 = g.threshold, distance: s2 = g.distance, includeMatches: i2 = g.includeMatches, findAllMatches: r2 = g.findAllMatches, minMatchCharLength: c2 = g.minMatchCharLength, isCaseSensitive: o2 = g.isCaseSensitive, ignoreLocation: h2 = g.ignoreLocation } = {}) {
    super(t2), this._bitapSearch = new L(t2, { location: e2, threshold: n2, distance: s2, includeMatches: i2, findAllMatches: r2, minMatchCharLength: c2, isCaseSensitive: o2, ignoreLocation: h2 });
  }
  static get type() {
    return "fuzzy";
  }
  static get multiRegex() {
    return /^"(.*)"$/;
  }
  static get singleRegex() {
    return /^(.*)$/;
  }
  search(t2) {
    return this._bitapSearch.searchIn(t2);
  }
};
var S = class extends k {
  constructor(t2) {
    super(t2);
  }
  static get type() {
    return "include";
  }
  static get multiRegex() {
    return /^'"(.*)"$/;
  }
  static get singleRegex() {
    return /^'(.*)$/;
  }
  search(t2) {
    let e2, n2 = 0;
    const s2 = [], i2 = this.pattern.length;
    for (; (e2 = t2.indexOf(this.pattern, n2)) > -1; )
      n2 = e2 + i2, s2.push([e2, n2 - 1]);
    const r2 = !!s2.length;
    return { isMatch: r2, score: r2 ? 0 : 1, indices: s2 };
  }
};
var w = [class extends k {
  constructor(t2) {
    super(t2);
  }
  static get type() {
    return "exact";
  }
  static get multiRegex() {
    return /^="(.*)"$/;
  }
  static get singleRegex() {
    return /^=(.*)$/;
  }
  search(t2) {
    const e2 = t2 === this.pattern;
    return { isMatch: e2, score: e2 ? 0 : 1, indices: [0, this.pattern.length - 1] };
  }
}, S, class extends k {
  constructor(t2) {
    super(t2);
  }
  static get type() {
    return "prefix-exact";
  }
  static get multiRegex() {
    return /^\^"(.*)"$/;
  }
  static get singleRegex() {
    return /^\^(.*)$/;
  }
  search(t2) {
    const e2 = t2.startsWith(this.pattern);
    return { isMatch: e2, score: e2 ? 0 : 1, indices: [0, this.pattern.length - 1] };
  }
}, class extends k {
  constructor(t2) {
    super(t2);
  }
  static get type() {
    return "inverse-prefix-exact";
  }
  static get multiRegex() {
    return /^!\^"(.*)"$/;
  }
  static get singleRegex() {
    return /^!\^(.*)$/;
  }
  search(t2) {
    const e2 = !t2.startsWith(this.pattern);
    return { isMatch: e2, score: e2 ? 0 : 1, indices: [0, t2.length - 1] };
  }
}, class extends k {
  constructor(t2) {
    super(t2);
  }
  static get type() {
    return "inverse-suffix-exact";
  }
  static get multiRegex() {
    return /^!"(.*)"\$$/;
  }
  static get singleRegex() {
    return /^!(.*)\$$/;
  }
  search(t2) {
    const e2 = !t2.endsWith(this.pattern);
    return { isMatch: e2, score: e2 ? 0 : 1, indices: [0, t2.length - 1] };
  }
}, class extends k {
  constructor(t2) {
    super(t2);
  }
  static get type() {
    return "suffix-exact";
  }
  static get multiRegex() {
    return /^"(.*)"\$$/;
  }
  static get singleRegex() {
    return /^(.*)\$$/;
  }
  search(t2) {
    const e2 = t2.endsWith(this.pattern);
    return { isMatch: e2, score: e2 ? 0 : 1, indices: [t2.length - this.pattern.length, t2.length - 1] };
  }
}, class extends k {
  constructor(t2) {
    super(t2);
  }
  static get type() {
    return "inverse-exact";
  }
  static get multiRegex() {
    return /^!"(.*)"$/;
  }
  static get singleRegex() {
    return /^!(.*)$/;
  }
  search(t2) {
    const e2 = -1 === t2.indexOf(this.pattern);
    return { isMatch: e2, score: e2 ? 0 : 1, indices: [0, t2.length - 1] };
  }
}, v];
var C = w.length;
var I = / +(?=(?:[^\"]*\"[^\"]*\")*[^\"]*$)/;
var $ = /* @__PURE__ */ new Set([v.type, S.type]);
var A = class {
  constructor(t2, { isCaseSensitive: e2 = g.isCaseSensitive, includeMatches: n2 = g.includeMatches, minMatchCharLength: s2 = g.minMatchCharLength, ignoreLocation: i2 = g.ignoreLocation, findAllMatches: r2 = g.findAllMatches, location: c2 = g.location, threshold: o2 = g.threshold, distance: h2 = g.distance } = {}) {
    this.query = null, this.options = { isCaseSensitive: e2, includeMatches: n2, minMatchCharLength: s2, findAllMatches: r2, ignoreLocation: i2, location: c2, threshold: o2, distance: h2 }, this.pattern = e2 ? t2 : t2.toLowerCase(), this.query = function(t3, e3 = {}) {
      return t3.split("|").map((t4) => {
        let n3 = t4.trim().split(I).filter((t5) => t5 && !!t5.trim()), s3 = [];
        for (let t5 = 0, i3 = n3.length; t5 < i3; t5 += 1) {
          const i4 = n3[t5];
          let r3 = false, c3 = -1;
          for (; !r3 && ++c3 < C; ) {
            const t6 = w[c3];
            let n4 = t6.isMultiMatch(i4);
            n4 && (s3.push(new t6(n4, e3)), r3 = true);
          }
          if (!r3)
            for (c3 = -1; ++c3 < C; ) {
              const t6 = w[c3];
              let n4 = t6.isSingleMatch(i4);
              if (n4) {
                s3.push(new t6(n4, e3));
                break;
              }
            }
        }
        return s3;
      });
    }(this.pattern, this.options);
  }
  static condition(t2, e2) {
    return e2.useExtendedSearch;
  }
  searchIn(t2) {
    const e2 = this.query;
    if (!e2)
      return { isMatch: false, score: 1 };
    const { includeMatches: n2, isCaseSensitive: s2 } = this.options;
    t2 = s2 ? t2 : t2.toLowerCase();
    let i2 = 0, r2 = [], c2 = 0;
    for (let s3 = 0, o2 = e2.length; s3 < o2; s3 += 1) {
      const o3 = e2[s3];
      r2.length = 0, i2 = 0;
      for (let e3 = 0, s4 = o3.length; e3 < s4; e3 += 1) {
        const s5 = o3[e3], { isMatch: h2, indices: a2, score: l2 } = s5.search(t2);
        if (!h2) {
          c2 = 0, i2 = 0, r2.length = 0;
          break;
        }
        if (i2 += 1, c2 += l2, n2) {
          const t3 = s5.constructor.type;
          $.has(t3) ? r2 = [...r2, ...a2] : r2.push(a2);
        }
      }
      if (i2) {
        let t3 = { isMatch: true, score: c2 / i2 };
        return n2 && (t3.indices = r2), t3;
      }
    }
    return { isMatch: false, score: 1 };
  }
};
var E = [];
function b(t2, e2) {
  for (let n2 = 0, s2 = E.length; n2 < s2; n2 += 1) {
    let s3 = E[n2];
    if (s3.condition(t2, e2))
      return new s3(t2, e2);
  }
  return new L(t2, e2);
}
var F = "$and";
var N = "$or";
var R = "$path";
var O = "$val";
var j = (t2) => !(!t2[F] && !t2[N]);
var W = (t2) => ({ [F]: Object.keys(t2).map((e2) => ({ [e2]: t2[e2] })) });
function z(n2, s2, { auto: r2 = true } = {}) {
  const c2 = (n3) => {
    let o2 = Object.keys(n3);
    const h2 = ((t2) => !!t2[R])(n3);
    if (!h2 && o2.length > 1 && !j(n3))
      return c2(W(n3));
    if (((e2) => !t(e2) && i(e2) && !j(e2))(n3)) {
      const t2 = h2 ? n3[R] : o2[0], i2 = h2 ? n3[O] : n3[t2];
      if (!e(i2))
        throw new Error(((t3) => `Invalid value for key ${t3}`)(t2));
      const c3 = { keyId: d(t2), pattern: i2 };
      return r2 && (c3.searcher = b(i2, s2)), c3;
    }
    let a2 = { children: [], operator: o2[0] };
    return o2.forEach((e2) => {
      const s3 = n3[e2];
      t(s3) && s3.forEach((t2) => {
        a2.children.push(c2(t2));
      });
    }), a2;
  };
  return j(n2) || (n2 = W(n2)), c2(n2);
}
function K(t2, e2) {
  const n2 = t2.matches;
  e2.matches = [], r(n2) && n2.forEach((t3) => {
    if (!r(t3.indices) || !t3.indices.length)
      return;
    const { indices: n3, value: s2 } = t3;
    let i2 = { indices: n3, value: s2 };
    t3.key && (i2.key = t3.key.src), t3.idx > -1 && (i2.refIndex = t3.idx), e2.matches.push(i2);
  });
}
function P(t2, e2) {
  e2.score = t2.score;
}
var q = class {
  constructor(t2, e2 = {}, n2) {
    this.options = { ...g, ...e2 }, this.options.useExtendedSearch, this._keyStore = new a(this.options.keys), this.setCollection(t2, n2);
  }
  setCollection(t2, e2) {
    if (this._docs = t2, e2 && !(e2 instanceof p))
      throw new Error("Incorrect 'index' type");
    this._myIndex = e2 || m(this.options.keys, this._docs, { getFn: this.options.getFn, fieldNormWeight: this.options.fieldNormWeight });
  }
  add(t2) {
    r(t2) && (this._docs.push(t2), this._myIndex.add(t2));
  }
  remove(t2 = () => false) {
    const e2 = [];
    for (let n2 = 0, s2 = this._docs.length; n2 < s2; n2 += 1) {
      const i2 = this._docs[n2];
      t2(i2, n2) && (this.removeAt(n2), n2 -= 1, s2 -= 1, e2.push(i2));
    }
    return e2;
  }
  removeAt(t2) {
    this._docs.splice(t2, 1), this._myIndex.removeAt(t2);
  }
  getIndex() {
    return this._myIndex;
  }
  search(t2, { limit: s2 = -1 } = {}) {
    const { includeMatches: i2, includeScore: r2, shouldSort: c2, sortFn: o2, ignoreFieldNorm: h2 } = this.options;
    let a2 = e(t2) ? e(this._docs[0]) ? this._searchStringList(t2) : this._searchObjectList(t2) : this._searchLogical(t2);
    return function(t3, { ignoreFieldNorm: e2 = g.ignoreFieldNorm }) {
      t3.forEach((t4) => {
        let n2 = 1;
        t4.matches.forEach(({ key: t5, norm: s3, score: i3 }) => {
          const r3 = t5 ? t5.weight : null;
          n2 *= Math.pow(0 === i3 && r3 ? Number.EPSILON : i3, (r3 || 1) * (e2 ? 1 : s3));
        }), t4.score = n2;
      });
    }(a2, { ignoreFieldNorm: h2 }), c2 && a2.sort(o2), n(s2) && s2 > -1 && (a2 = a2.slice(0, s2)), function(t3, e2, { includeMatches: n2 = g.includeMatches, includeScore: s3 = g.includeScore } = {}) {
      const i3 = [];
      return n2 && i3.push(K), s3 && i3.push(P), t3.map((t4) => {
        const { idx: n3 } = t4, s4 = { item: e2[n3], refIndex: n3 };
        return i3.length && i3.forEach((e3) => {
          e3(t4, s4);
        }), s4;
      });
    }(a2, this._docs, { includeMatches: i2, includeScore: r2 });
  }
  _searchStringList(t2) {
    const e2 = b(t2, this.options), { records: n2 } = this._myIndex, s2 = [];
    return n2.forEach(({ v: t3, i: n3, n: i2 }) => {
      if (!r(t3))
        return;
      const { isMatch: c2, score: o2, indices: h2 } = e2.searchIn(t3);
      c2 && s2.push({ item: t3, idx: n3, matches: [{ score: o2, value: t3, norm: i2, indices: h2 }] });
    }), s2;
  }
  _searchLogical(t2) {
    const e2 = z(t2, this.options), n2 = (t3, e3, s3) => {
      if (!t3.children) {
        const { keyId: n3, searcher: i4 } = t3, r2 = this._findMatches({ key: this._keyStore.get(n3), value: this._myIndex.getValueForItemAtKeyId(e3, n3), searcher: i4 });
        return r2 && r2.length ? [{ idx: s3, item: e3, matches: r2 }] : [];
      }
      const i3 = [];
      for (let r2 = 0, c3 = t3.children.length; r2 < c3; r2 += 1) {
        const c4 = t3.children[r2], o2 = n2(c4, e3, s3);
        if (o2.length)
          i3.push(...o2);
        else if (t3.operator === F)
          return [];
      }
      return i3;
    }, s2 = this._myIndex.records, i2 = {}, c2 = [];
    return s2.forEach(({ $: t3, i: s3 }) => {
      if (r(t3)) {
        let r2 = n2(e2, t3, s3);
        r2.length && (i2[s3] || (i2[s3] = { idx: s3, item: t3, matches: [] }, c2.push(i2[s3])), r2.forEach(({ matches: t4 }) => {
          i2[s3].matches.push(...t4);
        }));
      }
    }), c2;
  }
  _searchObjectList(t2) {
    const e2 = b(t2, this.options), { keys: n2, records: s2 } = this._myIndex, i2 = [];
    return s2.forEach(({ $: t3, i: s3 }) => {
      if (!r(t3))
        return;
      let c2 = [];
      n2.forEach((n3, s4) => {
        c2.push(...this._findMatches({ key: n3, value: t3[s4], searcher: e2 }));
      }), c2.length && i2.push({ idx: s3, item: t3, matches: c2 });
    }), i2;
  }
  _findMatches({ key: e2, value: n2, searcher: s2 }) {
    if (!r(n2))
      return [];
    let i2 = [];
    if (t(n2))
      n2.forEach(({ v: t2, i: n3, n: c2 }) => {
        if (!r(t2))
          return;
        const { isMatch: o2, score: h2, indices: a2 } = s2.searchIn(t2);
        o2 && i2.push({ score: h2, key: e2, value: t2, idx: n3, norm: c2, indices: a2 });
      });
    else {
      const { v: t2, n: r2 } = n2, { isMatch: c2, score: o2, indices: h2 } = s2.searchIn(t2);
      c2 && i2.push({ score: o2, key: e2, value: t2, norm: r2, indices: h2 });
    }
    return i2;
  }
};
q.version = "6.6.2", q.createIndex = m, q.parseIndex = function(t2, { getFn: e2 = g.getFn, fieldNormWeight: n2 = g.fieldNormWeight } = {}) {
  const { keys: s2, records: i2 } = t2, r2 = new p({ getFn: e2, fieldNormWeight: n2 });
  return r2.setKeys(s2), r2.setIndexRecords(i2), r2;
}, q.config = g, function(...t2) {
  E.push(...t2);
}(A);

// search/build.js
var myIndex = q.createIndex(["id", "title", "contents"], documents);
JSON.stringify(myIndex.toJSON());
