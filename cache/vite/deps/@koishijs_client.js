import {
  RouterLink,
  START_LOCATION_NORMALIZED,
  createRouter,
  createWebHistory
} from "./chunk-VCYEKLNT.js";
import {
  ElLoading,
  ElMessage,
  ElMessageBox,
  installer
} from "./chunk-NQ6HWGF3.js";
import {
  useLocalStorage,
  usePreferredDark
} from "./chunk-VCM64KFY.js";
import "./chunk-XR2QTC5I.js";
import {
  marked
} from "./chunk-AXCCOMI4.js";
import {
  require_lib
} from "./chunk-ESMDUVAB.js";
import {
  Binary,
  IconAdd,
  IconArrowDown,
  IconArrowUp,
  IconBranch,
  IconClose,
  IconCode,
  IconCollapse,
  IconDelete,
  IconEllipsis,
  IconExpand,
  IconExternal,
  IconEye,
  IconEyeSlash,
  IconInsertAfter,
  IconInsertBefore,
  IconInvalid,
  IconRedo,
  IconReset,
  IconSquareCheck,
  IconSquareEmpty,
  IconUndo,
  Primitive,
  Time,
  VirtualList,
  arrayBufferToBase64,
  arrayBufferToHex,
  base64ToArrayBuffer,
  camelCase,
  camelize,
  capitalize,
  client_default,
  clone,
  contain,
  deduplicate,
  deepEqual,
  defineProperty,
  difference,
  filterKeys,
  form,
  formatProperty,
  hexToArrayBuffer,
  hyphenate,
  intersection,
  is,
  isNonNullable,
  isNullable,
  isPlainObject,
  lib_default,
  makeArray,
  mapValues,
  noop,
  omit,
  paramCase,
  pick,
  remove,
  sanitize,
  snakeCase,
  src_default,
  trimSlash,
  uncapitalize,
  union,
  useI18nText
} from "./chunk-BZ4TRKRY.js";
import {
  DatetimeFormat,
  I18nD,
  I18nInjectionKey,
  I18nN,
  I18nT,
  NumberFormat,
  Translation,
  VERSION,
  castToVueI18n,
  createI18n,
  useI18n,
  vTDirective
} from "./chunk-OYU45GEP.js";
import "./chunk-5KG2YMYE.js";
import {
  computed,
  createApp,
  defineComponent,
  h,
  inject,
  markRaw,
  onBeforeUnmount,
  provide,
  reactive,
  ref,
  resolveComponent,
  shallowReactive,
  toValue,
  watch,
  watchEffect
} from "./chunk-QBIFJQ73.js";
import {
  __commonJS,
  __export,
  __publicField,
  __toESM
} from "./chunk-Q4XP6UTR.js";

// node_modules/.bun/supports-color@8.1.1/node_modules/supports-color/browser.js
var require_browser = __commonJS({
  "node_modules/.bun/supports-color@8.1.1/node_modules/supports-color/browser.js"(exports, module) {
    "use strict";
    function getChromeVersion() {
      const matches = /(Chrome|Chromium)\/(?<chromeVersion>\d+)\./.exec(navigator.userAgent);
      if (!matches) {
        return;
      }
      return Number.parseInt(matches.groups.chromeVersion, 10);
    }
    var colorSupport = getChromeVersion() >= 69 ? {
      level: 1,
      hasBasic: true,
      has256: false,
      has16m: false
    } : false;
    module.exports = {
      stdout: colorSupport,
      stderr: colorSupport
    };
  }
});

// (disabled):node_modules/.bun/object-inspect@1.13.4/node_modules/object-inspect/util.inspect
var require_util = __commonJS({
  "(disabled):node_modules/.bun/object-inspect@1.13.4/node_modules/object-inspect/util.inspect"() {
  }
});

// node_modules/.bun/object-inspect@1.13.4/node_modules/object-inspect/index.js
var require_object_inspect = __commonJS({
  "node_modules/.bun/object-inspect@1.13.4/node_modules/object-inspect/index.js"(exports, module) {
    var hasMap = typeof Map === "function" && Map.prototype;
    var mapSizeDescriptor = Object.getOwnPropertyDescriptor && hasMap ? Object.getOwnPropertyDescriptor(Map.prototype, "size") : null;
    var mapSize = hasMap && mapSizeDescriptor && typeof mapSizeDescriptor.get === "function" ? mapSizeDescriptor.get : null;
    var mapForEach = hasMap && Map.prototype.forEach;
    var hasSet = typeof Set === "function" && Set.prototype;
    var setSizeDescriptor = Object.getOwnPropertyDescriptor && hasSet ? Object.getOwnPropertyDescriptor(Set.prototype, "size") : null;
    var setSize = hasSet && setSizeDescriptor && typeof setSizeDescriptor.get === "function" ? setSizeDescriptor.get : null;
    var setForEach = hasSet && Set.prototype.forEach;
    var hasWeakMap = typeof WeakMap === "function" && WeakMap.prototype;
    var weakMapHas = hasWeakMap ? WeakMap.prototype.has : null;
    var hasWeakSet = typeof WeakSet === "function" && WeakSet.prototype;
    var weakSetHas = hasWeakSet ? WeakSet.prototype.has : null;
    var hasWeakRef = typeof WeakRef === "function" && WeakRef.prototype;
    var weakRefDeref = hasWeakRef ? WeakRef.prototype.deref : null;
    var booleanValueOf = Boolean.prototype.valueOf;
    var objectToString = Object.prototype.toString;
    var functionToString = Function.prototype.toString;
    var $match = String.prototype.match;
    var $slice = String.prototype.slice;
    var $replace = String.prototype.replace;
    var $toUpperCase = String.prototype.toUpperCase;
    var $toLowerCase = String.prototype.toLowerCase;
    var $test = RegExp.prototype.test;
    var $concat = Array.prototype.concat;
    var $join = Array.prototype.join;
    var $arrSlice = Array.prototype.slice;
    var $floor = Math.floor;
    var bigIntValueOf = typeof BigInt === "function" ? BigInt.prototype.valueOf : null;
    var gOPS = Object.getOwnPropertySymbols;
    var symToString = typeof Symbol === "function" && typeof Symbol.iterator === "symbol" ? Symbol.prototype.toString : null;
    var hasShammedSymbols = typeof Symbol === "function" && typeof Symbol.iterator === "object";
    var toStringTag = typeof Symbol === "function" && Symbol.toStringTag && (typeof Symbol.toStringTag === hasShammedSymbols ? "object" : "symbol") ? Symbol.toStringTag : null;
    var isEnumerable = Object.prototype.propertyIsEnumerable;
    var gPO = (typeof Reflect === "function" ? Reflect.getPrototypeOf : Object.getPrototypeOf) || ([].__proto__ === Array.prototype ? function(O) {
      return O.__proto__;
    } : null);
    function addNumericSeparator(num, str) {
      if (num === Infinity || num === -Infinity || num !== num || num && num > -1e3 && num < 1e3 || $test.call(/e/, str)) {
        return str;
      }
      var sepRegex = /[0-9](?=(?:[0-9]{3})+(?![0-9]))/g;
      if (typeof num === "number") {
        var int = num < 0 ? -$floor(-num) : $floor(num);
        if (int !== num) {
          var intStr = String(int);
          var dec = $slice.call(str, intStr.length + 1);
          return $replace.call(intStr, sepRegex, "$&_") + "." + $replace.call($replace.call(dec, /([0-9]{3})/g, "$&_"), /_$/, "");
        }
      }
      return $replace.call(str, sepRegex, "$&_");
    }
    var utilInspect = require_util();
    var inspectCustom = utilInspect.custom;
    var inspectSymbol = isSymbol(inspectCustom) ? inspectCustom : null;
    var quotes = {
      __proto__: null,
      "double": '"',
      single: "'"
    };
    var quoteREs = {
      __proto__: null,
      "double": /(["\\])/g,
      single: /(['\\])/g
    };
    module.exports = function inspect_(obj, options, depth, seen) {
      var opts = options || {};
      if (has(opts, "quoteStyle") && !has(quotes, opts.quoteStyle)) {
        throw new TypeError('option "quoteStyle" must be "single" or "double"');
      }
      if (has(opts, "maxStringLength") && (typeof opts.maxStringLength === "number" ? opts.maxStringLength < 0 && opts.maxStringLength !== Infinity : opts.maxStringLength !== null)) {
        throw new TypeError('option "maxStringLength", if provided, must be a positive integer, Infinity, or `null`');
      }
      var customInspect = has(opts, "customInspect") ? opts.customInspect : true;
      if (typeof customInspect !== "boolean" && customInspect !== "symbol") {
        throw new TypeError("option \"customInspect\", if provided, must be `true`, `false`, or `'symbol'`");
      }
      if (has(opts, "indent") && opts.indent !== null && opts.indent !== "	" && !(parseInt(opts.indent, 10) === opts.indent && opts.indent > 0)) {
        throw new TypeError('option "indent" must be "\\t", an integer > 0, or `null`');
      }
      if (has(opts, "numericSeparator") && typeof opts.numericSeparator !== "boolean") {
        throw new TypeError('option "numericSeparator", if provided, must be `true` or `false`');
      }
      var numericSeparator = opts.numericSeparator;
      if (typeof obj === "undefined") {
        return "undefined";
      }
      if (obj === null) {
        return "null";
      }
      if (typeof obj === "boolean") {
        return obj ? "true" : "false";
      }
      if (typeof obj === "string") {
        return inspectString(obj, opts);
      }
      if (typeof obj === "number") {
        if (obj === 0) {
          return Infinity / obj > 0 ? "0" : "-0";
        }
        var str = String(obj);
        return numericSeparator ? addNumericSeparator(obj, str) : str;
      }
      if (typeof obj === "bigint") {
        var bigIntStr = String(obj) + "n";
        return numericSeparator ? addNumericSeparator(obj, bigIntStr) : bigIntStr;
      }
      var maxDepth = typeof opts.depth === "undefined" ? 5 : opts.depth;
      if (typeof depth === "undefined") {
        depth = 0;
      }
      if (depth >= maxDepth && maxDepth > 0 && typeof obj === "object") {
        return isArray(obj) ? "[Array]" : "[Object]";
      }
      var indent = getIndent(opts, depth);
      if (typeof seen === "undefined") {
        seen = [];
      } else if (indexOf(seen, obj) >= 0) {
        return "[Circular]";
      }
      function inspect2(value, from, noIndent) {
        if (from) {
          seen = $arrSlice.call(seen);
          seen.push(from);
        }
        if (noIndent) {
          var newOpts = {
            depth: opts.depth
          };
          if (has(opts, "quoteStyle")) {
            newOpts.quoteStyle = opts.quoteStyle;
          }
          return inspect_(value, newOpts, depth + 1, seen);
        }
        return inspect_(value, opts, depth + 1, seen);
      }
      if (typeof obj === "function" && !isRegExp(obj)) {
        var name = nameOf(obj);
        var keys = arrObjKeys(obj, inspect2);
        return "[Function" + (name ? ": " + name : " (anonymous)") + "]" + (keys.length > 0 ? " { " + $join.call(keys, ", ") + " }" : "");
      }
      if (isSymbol(obj)) {
        var symString = hasShammedSymbols ? $replace.call(String(obj), /^(Symbol\(.*\))_[^)]*$/, "$1") : symToString.call(obj);
        return typeof obj === "object" && !hasShammedSymbols ? markBoxed(symString) : symString;
      }
      if (isElement(obj)) {
        var s = "<" + $toLowerCase.call(String(obj.nodeName));
        var attrs = obj.attributes || [];
        for (var i = 0; i < attrs.length; i++) {
          s += " " + attrs[i].name + "=" + wrapQuotes(quote(attrs[i].value), "double", opts);
        }
        s += ">";
        if (obj.childNodes && obj.childNodes.length) {
          s += "...";
        }
        s += "</" + $toLowerCase.call(String(obj.nodeName)) + ">";
        return s;
      }
      if (isArray(obj)) {
        if (obj.length === 0) {
          return "[]";
        }
        var xs = arrObjKeys(obj, inspect2);
        if (indent && !singleLineValues(xs)) {
          return "[" + indentedJoin(xs, indent) + "]";
        }
        return "[ " + $join.call(xs, ", ") + " ]";
      }
      if (isError(obj)) {
        var parts = arrObjKeys(obj, inspect2);
        if (!("cause" in Error.prototype) && "cause" in obj && !isEnumerable.call(obj, "cause")) {
          return "{ [" + String(obj) + "] " + $join.call($concat.call("[cause]: " + inspect2(obj.cause), parts), ", ") + " }";
        }
        if (parts.length === 0) {
          return "[" + String(obj) + "]";
        }
        return "{ [" + String(obj) + "] " + $join.call(parts, ", ") + " }";
      }
      if (typeof obj === "object" && customInspect) {
        if (inspectSymbol && typeof obj[inspectSymbol] === "function" && utilInspect) {
          return utilInspect(obj, { depth: maxDepth - depth });
        } else if (customInspect !== "symbol" && typeof obj.inspect === "function") {
          return obj.inspect();
        }
      }
      if (isMap(obj)) {
        var mapParts = [];
        if (mapForEach) {
          mapForEach.call(obj, function(value, key) {
            mapParts.push(inspect2(key, obj, true) + " => " + inspect2(value, obj));
          });
        }
        return collectionOf("Map", mapSize.call(obj), mapParts, indent);
      }
      if (isSet(obj)) {
        var setParts = [];
        if (setForEach) {
          setForEach.call(obj, function(value) {
            setParts.push(inspect2(value, obj));
          });
        }
        return collectionOf("Set", setSize.call(obj), setParts, indent);
      }
      if (isWeakMap(obj)) {
        return weakCollectionOf("WeakMap");
      }
      if (isWeakSet(obj)) {
        return weakCollectionOf("WeakSet");
      }
      if (isWeakRef(obj)) {
        return weakCollectionOf("WeakRef");
      }
      if (isNumber(obj)) {
        return markBoxed(inspect2(Number(obj)));
      }
      if (isBigInt(obj)) {
        return markBoxed(inspect2(bigIntValueOf.call(obj)));
      }
      if (isBoolean(obj)) {
        return markBoxed(booleanValueOf.call(obj));
      }
      if (isString(obj)) {
        return markBoxed(inspect2(String(obj)));
      }
      if (typeof window !== "undefined" && obj === window) {
        return "{ [object Window] }";
      }
      if (typeof globalThis !== "undefined" && obj === globalThis || typeof global !== "undefined" && obj === global) {
        return "{ [object globalThis] }";
      }
      if (!isDate(obj) && !isRegExp(obj)) {
        var ys = arrObjKeys(obj, inspect2);
        var isPlainObject2 = gPO ? gPO(obj) === Object.prototype : obj instanceof Object || obj.constructor === Object;
        var protoTag = obj instanceof Object ? "" : "null prototype";
        var stringTag = !isPlainObject2 && toStringTag && Object(obj) === obj && toStringTag in obj ? $slice.call(toStr(obj), 8, -1) : protoTag ? "Object" : "";
        var constructorTag = isPlainObject2 || typeof obj.constructor !== "function" ? "" : obj.constructor.name ? obj.constructor.name + " " : "";
        var tag = constructorTag + (stringTag || protoTag ? "[" + $join.call($concat.call([], stringTag || [], protoTag || []), ": ") + "] " : "");
        if (ys.length === 0) {
          return tag + "{}";
        }
        if (indent) {
          return tag + "{" + indentedJoin(ys, indent) + "}";
        }
        return tag + "{ " + $join.call(ys, ", ") + " }";
      }
      return String(obj);
    };
    function wrapQuotes(s, defaultStyle, opts) {
      var style = opts.quoteStyle || defaultStyle;
      var quoteChar = quotes[style];
      return quoteChar + s + quoteChar;
    }
    function quote(s) {
      return $replace.call(String(s), /"/g, "&quot;");
    }
    function canTrustToString(obj) {
      return !toStringTag || !(typeof obj === "object" && (toStringTag in obj || typeof obj[toStringTag] !== "undefined"));
    }
    function isArray(obj) {
      return toStr(obj) === "[object Array]" && canTrustToString(obj);
    }
    function isDate(obj) {
      return toStr(obj) === "[object Date]" && canTrustToString(obj);
    }
    function isRegExp(obj) {
      return toStr(obj) === "[object RegExp]" && canTrustToString(obj);
    }
    function isError(obj) {
      return toStr(obj) === "[object Error]" && canTrustToString(obj);
    }
    function isString(obj) {
      return toStr(obj) === "[object String]" && canTrustToString(obj);
    }
    function isNumber(obj) {
      return toStr(obj) === "[object Number]" && canTrustToString(obj);
    }
    function isBoolean(obj) {
      return toStr(obj) === "[object Boolean]" && canTrustToString(obj);
    }
    function isSymbol(obj) {
      if (hasShammedSymbols) {
        return obj && typeof obj === "object" && obj instanceof Symbol;
      }
      if (typeof obj === "symbol") {
        return true;
      }
      if (!obj || typeof obj !== "object" || !symToString) {
        return false;
      }
      try {
        symToString.call(obj);
        return true;
      } catch (e) {
      }
      return false;
    }
    function isBigInt(obj) {
      if (!obj || typeof obj !== "object" || !bigIntValueOf) {
        return false;
      }
      try {
        bigIntValueOf.call(obj);
        return true;
      } catch (e) {
      }
      return false;
    }
    var hasOwn = Object.prototype.hasOwnProperty || function(key) {
      return key in this;
    };
    function has(obj, key) {
      return hasOwn.call(obj, key);
    }
    function toStr(obj) {
      return objectToString.call(obj);
    }
    function nameOf(f) {
      if (f.name) {
        return f.name;
      }
      var m = $match.call(functionToString.call(f), /^function\s*([\w$]+)/);
      if (m) {
        return m[1];
      }
      return null;
    }
    function indexOf(xs, x) {
      if (xs.indexOf) {
        return xs.indexOf(x);
      }
      for (var i = 0, l = xs.length; i < l; i++) {
        if (xs[i] === x) {
          return i;
        }
      }
      return -1;
    }
    function isMap(x) {
      if (!mapSize || !x || typeof x !== "object") {
        return false;
      }
      try {
        mapSize.call(x);
        try {
          setSize.call(x);
        } catch (s) {
          return true;
        }
        return x instanceof Map;
      } catch (e) {
      }
      return false;
    }
    function isWeakMap(x) {
      if (!weakMapHas || !x || typeof x !== "object") {
        return false;
      }
      try {
        weakMapHas.call(x, weakMapHas);
        try {
          weakSetHas.call(x, weakSetHas);
        } catch (s) {
          return true;
        }
        return x instanceof WeakMap;
      } catch (e) {
      }
      return false;
    }
    function isWeakRef(x) {
      if (!weakRefDeref || !x || typeof x !== "object") {
        return false;
      }
      try {
        weakRefDeref.call(x);
        return true;
      } catch (e) {
      }
      return false;
    }
    function isSet(x) {
      if (!setSize || !x || typeof x !== "object") {
        return false;
      }
      try {
        setSize.call(x);
        try {
          mapSize.call(x);
        } catch (m) {
          return true;
        }
        return x instanceof Set;
      } catch (e) {
      }
      return false;
    }
    function isWeakSet(x) {
      if (!weakSetHas || !x || typeof x !== "object") {
        return false;
      }
      try {
        weakSetHas.call(x, weakSetHas);
        try {
          weakMapHas.call(x, weakMapHas);
        } catch (s) {
          return true;
        }
        return x instanceof WeakSet;
      } catch (e) {
      }
      return false;
    }
    function isElement(x) {
      if (!x || typeof x !== "object") {
        return false;
      }
      if (typeof HTMLElement !== "undefined" && x instanceof HTMLElement) {
        return true;
      }
      return typeof x.nodeName === "string" && typeof x.getAttribute === "function";
    }
    function inspectString(str, opts) {
      if (str.length > opts.maxStringLength) {
        var remaining = str.length - opts.maxStringLength;
        var trailer = "... " + remaining + " more character" + (remaining > 1 ? "s" : "");
        return inspectString($slice.call(str, 0, opts.maxStringLength), opts) + trailer;
      }
      var quoteRE = quoteREs[opts.quoteStyle || "single"];
      quoteRE.lastIndex = 0;
      var s = $replace.call($replace.call(str, quoteRE, "\\$1"), /[\x00-\x1f]/g, lowbyte);
      return wrapQuotes(s, "single", opts);
    }
    function lowbyte(c) {
      var n = c.charCodeAt(0);
      var x = {
        8: "b",
        9: "t",
        10: "n",
        12: "f",
        13: "r"
      }[n];
      if (x) {
        return "\\" + x;
      }
      return "\\x" + (n < 16 ? "0" : "") + $toUpperCase.call(n.toString(16));
    }
    function markBoxed(str) {
      return "Object(" + str + ")";
    }
    function weakCollectionOf(type) {
      return type + " { ? }";
    }
    function collectionOf(type, size, entries, indent) {
      var joinedEntries = indent ? indentedJoin(entries, indent) : $join.call(entries, ", ");
      return type + " (" + size + ") {" + joinedEntries + "}";
    }
    function singleLineValues(xs) {
      for (var i = 0; i < xs.length; i++) {
        if (indexOf(xs[i], "\n") >= 0) {
          return false;
        }
      }
      return true;
    }
    function getIndent(opts, depth) {
      var baseIndent;
      if (opts.indent === "	") {
        baseIndent = "	";
      } else if (typeof opts.indent === "number" && opts.indent > 0) {
        baseIndent = $join.call(Array(opts.indent + 1), " ");
      } else {
        return null;
      }
      return {
        base: baseIndent,
        prev: $join.call(Array(depth + 1), baseIndent)
      };
    }
    function indentedJoin(xs, indent) {
      if (xs.length === 0) {
        return "";
      }
      var lineJoiner = "\n" + indent.prev + indent.base;
      return lineJoiner + $join.call(xs, "," + lineJoiner) + "\n" + indent.prev;
    }
    function arrObjKeys(obj, inspect2) {
      var isArr = isArray(obj);
      var xs = [];
      if (isArr) {
        xs.length = obj.length;
        for (var i = 0; i < obj.length; i++) {
          xs[i] = has(obj, i) ? inspect2(obj[i], obj) : "";
        }
      }
      var syms = typeof gOPS === "function" ? gOPS(obj) : [];
      var symMap;
      if (hasShammedSymbols) {
        symMap = {};
        for (var k = 0; k < syms.length; k++) {
          symMap["$" + syms[k]] = syms[k];
        }
      }
      for (var key in obj) {
        if (!has(obj, key)) {
          continue;
        }
        if (isArr && String(Number(key)) === key && key < obj.length) {
          continue;
        }
        if (hasShammedSymbols && symMap["$" + key] instanceof Symbol) {
          continue;
        } else if ($test.call(/[^\w$]/, key)) {
          xs.push(inspect2(key, obj) + ": " + inspect2(obj[key], obj));
        } else {
          xs.push(key + ": " + inspect2(obj[key], obj));
        }
      }
      if (typeof gOPS === "function") {
        for (var j = 0; j < syms.length; j++) {
          if (isEnumerable.call(obj, syms[j])) {
            xs.push("[" + inspect2(syms[j]) + "]: " + inspect2(obj[syms[j]], obj));
          }
        }
      }
      return xs;
    }
  }
});

// node_modules/.bun/marked-vue@1.3.0+cc1f68d090f48582/node_modules/marked-vue/lib/index.mjs
var xss = __toESM(require_lib(), 1);
var __defProp = Object.defineProperty;
var __name = (target, value) => __defProp(target, "name", { value, configurable: true });
var allowedTags = [
  // Content sectioning
  "address",
  "article",
  "aside",
  "footer",
  "header",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "hgroup",
  "main",
  "nav",
  "section",
  // Text content
  "blockquote",
  "dd",
  "div",
  "dl",
  "dt",
  "figcaption",
  "figure",
  "hr",
  "li",
  "main",
  "ol",
  "p",
  "pre",
  "ul",
  // Inline text semantics
  "a",
  "abbr",
  "b",
  "bdi",
  "bdo",
  "br",
  "cite",
  "code",
  "data",
  "dfn",
  "em",
  "i",
  "kbd",
  "mark",
  "q",
  "rb",
  "rp",
  "rt",
  "rtc",
  "ruby",
  "s",
  "samp",
  "small",
  "span",
  "strong",
  "sub",
  "sup",
  "time",
  "u",
  "var",
  "wbr",
  // Table content
  "caption",
  "col",
  "colgroup",
  "table",
  "tbody",
  "td",
  "tfoot",
  "th",
  "thead",
  "tr"
];
var voidTags = ["img", "br", "hr", "area", "base", "basefont", "input", "link", "meta"];
var allowedProtocols = ["http:", "https:", "mailto:", "tel:"];
function checkUrl(value) {
  try {
    const url = new URL(value, location.toString());
    return allowedProtocols.includes(url.protocol);
  } catch (e) {
    return false;
  }
}
__name(checkUrl, "checkUrl");
function sanitize2(html) {
  const whiteList = {
    ...Object.fromEntries(allowedTags.map((tag) => [tag, []]))
  };
  const stack = [];
  html = xss.filterXSS(html, {
    whiteList,
    stripIgnoreTag: true,
    onTag(tag, raw, options) {
      let html2;
      if (tag === "a" && !options.isClosing) {
        const attrs = {};
        xss.parseAttr(raw.slice(3), (name, value) => {
          if (name === "href") {
            attrs[name] = checkUrl(value) ? value : "#";
          } else if (name === "title") {
            attrs[name] = xss.escapeAttrValue(value);
          }
          return "";
        });
        attrs.rel = "noopener noreferrer";
        attrs.target = "_blank";
        html2 = `<a ${Object.entries(attrs).map(([name, value]) => `${name}="${value}"`).join(" ")}>`;
      }
      if (raw.endsWith("/>") || voidTags.includes(tag))
        return;
      if (!options.isClosing) {
        stack.push(tag);
        return html2;
      }
      let result = "";
      while (stack.length) {
        const last = stack.pop();
        if (last === tag) {
          return result + raw;
        }
        result += `</${last}>`;
      }
      return raw.replace(/</g, "&lt;").replace(/>/g, "&gt;");
    }
  });
  while (stack.length) {
    const last = stack.pop();
    html += `</${last}>`;
  }
  return html;
}
__name(sanitize2, "sanitize");
var src_default2 = defineComponent({
  props: {
    source: String,
    inline: Boolean,
    tag: String,
    unsafe: Boolean
  },
  setup(props) {
    return () => {
      let html = props.inline ? marked.parseInline(props.source || "") : marked.parse(props.source || "");
      if (!props.unsafe)
        html = sanitize2(html);
      const tag = props.tag || (props.inline ? "span" : "div");
      return h(tag, {
        class: "markdown",
        innerHTML: html
      });
    };
  }
});

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/data.ts
var global2 = KOISHI_CONFIG;
var store = reactive({});
function withProxy(url) {
  return (global2.proxyBase || "") + url;
}
var socket = ref(null);
var listeners = {};
var responseHooks = {};
function send(type, ...args) {
  if (!socket.value) return;
  console.debug("↑%c", "color:brown", type, args);
  const id = Math.random().toString(36).slice(2, 9);
  socket.value.send(JSON.stringify({ id, type, args }));
  return new Promise((resolve, reject) => {
    responseHooks[id] = [resolve, reject];
    setTimeout(() => {
      delete responseHooks[id];
      reject(new Error("timeout"));
    }, 6e4);
  });
}
function receive(event, listener) {
  listeners[event] = listener;
}
receive("data", ({ key, value }) => {
  store[key] = value;
});
receive("patch", ({ key, value }) => {
  if (Array.isArray(store[key])) {
    store[key].push(...value);
  } else if (store[key]) {
    Object.assign(store[key], value);
  }
});
receive("response", ({ id, value, error }) => {
  if (!responseHooks[id]) return;
  const [resolve, reject] = responseHooks[id];
  delete responseHooks[id];
  if (error) {
    reject(error);
  } else {
    resolve(value);
  }
});
function connect(ctx, callback) {
  const value = callback();
  let sendTimer;
  let closeTimer;
  const refresh = () => {
    if (!global2.heartbeat) return;
    clearTimeout(sendTimer);
    clearTimeout(closeTimer);
    sendTimer = +setTimeout(() => send("ping"), global2.heartbeat.interval);
    closeTimer = +setTimeout(() => value == null ? void 0 : value.close(), global2.heartbeat.timeout);
  };
  const reconnect = () => {
    socket.value = null;
    for (const key in store) {
      store[key] = void 0;
    }
    console.log("[koishi] websocket disconnected, will retry in 1s...");
    setTimeout(() => {
      connect(ctx, callback).then(location.reload, () => {
        console.log("[koishi] websocket disconnected, will retry in 1s...");
      });
    }, 1e3);
  };
  value.addEventListener("message", (ev) => {
    refresh();
    const data = JSON.parse(ev.data);
    console.debug("↓%c", "color:purple", data.type, data.body);
    if (data.type in listeners) {
      listeners[data.type](data.body);
    }
    ctx.emit(data.type, data.body);
  });
  value.addEventListener("close", reconnect);
  return new Promise((resolve, reject) => {
    value.addEventListener("open", (event) => {
      socket.value = markRaw(value);
      resolve(event);
    });
    value.addEventListener("error", reject);
  });
}

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/common/index.ts
import Button from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/common/k-button.vue";
import Hint from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/common/k-hint.vue";
import Tab from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/common/k-tab.vue";
function common_default(app) {
  app.component("k-button", Button);
  app.component("k-hint", Hint);
  app.component("k-tab", Tab);
}

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/index.ts
import Dynamic from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/dynamic.vue";
import Perms from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/perms.vue";
import ChatImage from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/chat/image.vue";

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/index.ts
var icons_exports = {};
__export(icons_exports, {
  install: () => install,
  register: () => register
});
import Default from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/activity/default.vue";
import Ellipsis from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/activity/ellipsis.vue";
import Home from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/activity/home.vue";
import Moon from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/activity/moon.vue";
import Settings from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/activity/settings.vue";
import Sun from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/activity/sun.vue";
import ArrowLeft from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/arrow-left.vue";
import ArrowRight from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/arrow-right.vue";
import BoxOpen from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/box-open.vue";
import CheckFull from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/check-full.vue";
import ChevronDown from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/chevron-down.vue";
import ChevronLeft from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/chevron-left.vue";
import ChevronRight from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/chevron-right.vue";
import ChevronUp from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/chevron-up.vue";
import ClipboardList from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/clipboard-list.vue";
import Edit from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/edit.vue";
import ExclamationFull from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/exclamation-full.vue";
import Expand from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/expand.vue";
import FileArchive from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/file-archive.vue";
import Filter from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/filter.vue";
import GitHub from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/github.vue";
import GitLab from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/gitlab.vue";
import InfoFull from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/info-full.vue";
import Koishi from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/koishi.vue";
import Link from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/link.vue";
import PaperPlane from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/paper-plane.vue";
import QuestionEmpty from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/question-empty.vue";
import Redo from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/redo.vue";
import Search from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/search.vue";
import SearchMinus from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/search-minus.vue";
import SearchPlus from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/search-plus.vue";
import StarEmpty from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/star-empty.vue";
import StarFull from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/star-full.vue";
import Start from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/start.vue";
import Tag from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/tag.vue";
import TimesFull from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/times-full.vue";
import Tools from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/tools.vue";
import Undo from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/undo.vue";
import User from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/svg/user.vue";
import "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/icons/style.scss";
var registry = reactive({});
register("activity:default", Default);
register("activity:ellipsis", Ellipsis);
register("activity:home", Home);
register("activity:moon", Moon);
register("activity:settings", Settings);
register("activity:sun", Sun);
register("arrow-up", IconArrowUp);
register("arrow-down", IconArrowDown);
register("arrow-left", ArrowLeft);
register("arrow-right", ArrowRight);
register("box-open", BoxOpen);
register("check-full", CheckFull);
register("chevron-down", ChevronDown);
register("chevron-left", ChevronLeft);
register("chevron-right", ChevronRight);
register("chevron-up", ChevronUp);
register("clipboard-list", ClipboardList);
register("close", IconClose);
register("delete", IconDelete);
register("edit", Edit);
register("ellipsis", IconEllipsis);
register("exclamation-full", ExclamationFull);
register("expand", Expand);
register("external", IconExternal);
register("eye-slash", IconEyeSlash);
register("eye", IconEye);
register("file-archive", FileArchive);
register("filter", Filter);
register("github", GitHub);
register("gitlab", GitLab);
register("info-full", InfoFull);
register("koishi", Koishi);
register("link", Link);
register("paper-plane", PaperPlane);
register("add", IconAdd);
register("question-empty", QuestionEmpty);
register("redo", Redo);
register("search", Search);
register("search-minus", SearchMinus);
register("search-plus", SearchPlus);
register("star-empty", StarEmpty);
register("star-full", StarFull);
register("start", Start);
register("tag", Tag);
register("times-full", TimesFull);
register("tools", Tools);
register("undo", Undo);
register("user", User);
function register(name, component) {
  registry[name] = markRaw(component);
}
function install(app) {
  app.component("k-icon", defineComponent({
    props: {
      name: String
    },
    render(props) {
      const component = registry[props.name];
      return component && h(component);
    }
  }));
}

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/layout/index.ts
import Card from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/layout/card.vue";
import Content from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/layout/content.vue";
import Empty from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/layout/empty.vue";
import TabGroup from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/layout/tab-group.vue";
import TabItem from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/layout/tab-item.vue";
function layout_default(app) {
  app.component("k-card", Card);
  app.component("k-content", Content);
  app.component("k-empty", Empty);
  app.component("k-tab-group", TabGroup);
  app.component("k-tab-item", TabItem);
}

// node_modules/.bun/@cordisjs+core@3.18.1/node_modules/@cordisjs/core/lib/index.mjs
var __defProp2 = Object.defineProperty;
var __name2 = (target, value) => __defProp2(target, "name", { value, configurable: true });
var symbols = {
  // internal symbols
  shadow: Symbol.for("cordis.shadow"),
  receiver: Symbol.for("cordis.receiver"),
  original: Symbol.for("cordis.original"),
  // context symbols
  store: Symbol.for("cordis.store"),
  events: Symbol.for("cordis.events"),
  static: Symbol.for("cordis.static"),
  filter: Symbol.for("cordis.filter"),
  expose: Symbol.for("cordis.expose"),
  isolate: Symbol.for("cordis.isolate"),
  internal: Symbol.for("cordis.internal"),
  intercept: Symbol.for("cordis.intercept"),
  // service symbols
  setup: Symbol.for("cordis.setup"),
  invoke: Symbol.for("cordis.invoke"),
  extend: Symbol.for("cordis.extend"),
  tracker: Symbol.for("cordis.tracker"),
  provide: Symbol.for("cordis.provide"),
  immediate: Symbol.for("cordis.immediate")
};
var GeneratorFunction = (function* () {
}).constructor;
var AsyncGeneratorFunction = (async function* () {
}).constructor;
function isConstructor(func) {
  if (!func.prototype) return false;
  if (func instanceof GeneratorFunction) return false;
  if (AsyncGeneratorFunction !== Function && func instanceof AsyncGeneratorFunction) return false;
  return true;
}
__name2(isConstructor, "isConstructor");
function resolveConfig(plugin, config3) {
  const schema = plugin["Config"] || plugin["schema"];
  if (schema && plugin["schema"] !== false) config3 = schema(config3);
  return config3 ?? {};
}
__name2(resolveConfig, "resolveConfig");
function isUnproxyable(value) {
  return [Map, Set, Date, Promise].some((constructor) => value instanceof constructor);
}
__name2(isUnproxyable, "isUnproxyable");
function joinPrototype(proto1, proto2) {
  if (proto1 === Object.prototype) return proto2;
  const result = Object.create(joinPrototype(Object.getPrototypeOf(proto1), proto2));
  for (const key of Reflect.ownKeys(proto1)) {
    Object.defineProperty(result, key, Object.getOwnPropertyDescriptor(proto1, key));
  }
  return result;
}
__name2(joinPrototype, "joinPrototype");
function isObject(value) {
  return value && (typeof value === "object" || typeof value === "function");
}
__name2(isObject, "isObject");
function getTraceable(ctx, value, noTrap) {
  if (!isObject(value)) return value;
  if (Object.hasOwn(value, symbols.shadow)) {
    return Object.getPrototypeOf(value);
  }
  const tracker = value[symbols.tracker];
  if (!tracker) return value;
  return createTraceable(ctx, value, tracker, noTrap);
}
__name2(getTraceable, "getTraceable");
function withProps(target, props) {
  if (!props) return target;
  return new Proxy(target, {
    get: __name2((target2, prop, receiver) => {
      if (prop in props && prop !== "constructor") return Reflect.get(props, prop, receiver);
      return Reflect.get(target2, prop, receiver);
    }, "get"),
    set: __name2((target2, prop, value, receiver) => {
      if (prop in props && prop !== "constructor") return Reflect.set(props, prop, value, receiver);
      return Reflect.set(target2, prop, value, receiver);
    }, "set")
  });
}
__name2(withProps, "withProps");
function withProp(target, prop, value) {
  return withProps(target, Object.defineProperty(/* @__PURE__ */ Object.create(null), prop, {
    value,
    writable: false
  }));
}
__name2(withProp, "withProp");
function createShadow(ctx, target, property, receiver) {
  var _a15;
  if (!property) return receiver;
  const origin = (_a15 = Reflect.getOwnPropertyDescriptor(target, property)) == null ? void 0 : _a15.value;
  if (!origin) return receiver;
  return withProp(receiver, property, ctx.extend({ [symbols.shadow]: origin }));
}
__name2(createShadow, "createShadow");
function createShadowMethod(ctx, value, outer, shadow) {
  return new Proxy(value, {
    apply: __name2((target, thisArg, args) => {
      if (thisArg === outer) thisArg = shadow;
      args = args.map((arg) => {
        if (typeof arg !== "function" || arg[symbols.original]) return arg;
        return new Proxy(arg, {
          get: __name2((target2, prop, receiver) => {
            if (prop === symbols.original) return target2;
            const value2 = Reflect.get(target2, prop, receiver);
            if (prop === "toString" && value2 === Function.prototype.toString) {
              return function(...args2) {
                return Reflect.apply(value2, this === receiver ? target2 : this, args2);
              };
            }
            return value2;
          }, "get"),
          apply: __name2((target2, thisArg2, args2) => {
            return Reflect.apply(target2, getTraceable(ctx, thisArg2), args2.map((arg2) => getTraceable(ctx, arg2)));
          }, "apply"),
          construct: __name2((target2, args2, newTarget) => {
            return Reflect.construct(target2, args2.map((arg2) => getTraceable(ctx, arg2)), newTarget);
          }, "construct")
        });
      });
      return getTraceable(ctx, Reflect.apply(target, thisArg, args));
    }, "apply")
  });
}
__name2(createShadowMethod, "createShadowMethod");
function createTraceable(ctx, value, tracker, noTrap) {
  if (ctx[symbols.shadow]) {
    ctx = Object.getPrototypeOf(ctx);
  }
  const proxy = new Proxy(value, {
    get: __name2((target, prop, receiver) => {
      if (prop === symbols.original) return target;
      if (prop === tracker.property) return ctx;
      if (typeof prop === "symbol") {
        return Reflect.get(target, prop, receiver);
      }
      if (tracker.associate && ctx[symbols.internal][`${tracker.associate}.${prop}`]) {
        return Reflect.get(ctx, `${tracker.associate}.${prop}`, withProp(ctx, symbols.receiver, receiver));
      }
      const shadow = createShadow(ctx, target, tracker.property, receiver);
      const innerValue = Reflect.get(target, prop, shadow);
      const innerTracker = innerValue == null ? void 0 : innerValue[symbols.tracker];
      if (innerTracker) {
        return createTraceable(ctx, innerValue, innerTracker);
      } else if (!noTrap && typeof innerValue === "function") {
        return createShadowMethod(ctx, innerValue, receiver, shadow);
      } else {
        return innerValue;
      }
    }, "get"),
    set: __name2((target, prop, value2, receiver) => {
      if (prop === symbols.original) return false;
      if (prop === tracker.property) return false;
      if (typeof prop === "symbol") {
        return Reflect.set(target, prop, value2, receiver);
      }
      if (tracker.associate && ctx[symbols.internal][`${tracker.associate}.${prop}`]) {
        return Reflect.set(ctx, `${tracker.associate}.${prop}`, value2, withProp(ctx, symbols.receiver, receiver));
      }
      const shadow = createShadow(ctx, target, tracker.property, receiver);
      return Reflect.set(target, prop, value2, shadow);
    }, "set"),
    apply: __name2((target, thisArg, args) => {
      return applyTraceable(proxy, target, thisArg, args);
    }, "apply")
  });
  return proxy;
}
__name2(createTraceable, "createTraceable");
function applyTraceable(proxy, value, thisArg, args) {
  if (!value[symbols.invoke]) return Reflect.apply(value, thisArg, args);
  return value[symbols.invoke].apply(proxy, args);
}
__name2(applyTraceable, "applyTraceable");
function createCallable(name, proto, tracker) {
  const self = __name2(function(...args) {
    const proxy = createTraceable(self["ctx"], self, tracker);
    return applyTraceable(proxy, self, this, args);
  }, "self");
  defineProperty(self, "name", name);
  return Object.setPrototypeOf(self, proto);
}
__name2(createCallable, "createCallable");
var _a;
var ReflectService = (_a = class {
  constructor(ctx) {
    this.ctx = ctx;
    defineProperty(this, symbols.tracker, {
      associate: "reflect",
      property: "ctx"
    });
    this._mixin("reflect", ["get", "set", "provide", "accessor", "mixin", "alias"]);
    this._mixin("scope", ["config", "runtime", "effect", "collect", "accept", "decline"]);
    this._mixin("registry", ["using", "inject", "plugin"]);
    this._mixin("lifecycle", ["on", "once", "parallel", "emit", "serial", "bail", "start", "stop"]);
  }
  static resolveInject(ctx, name) {
    let internal = ctx[symbols.internal][name];
    while ((internal == null ? void 0 : internal.type) === "alias") {
      name = internal.name;
      internal = ctx[symbols.internal][name];
    }
    return [name, internal];
  }
  static checkInject(ctx, name, error) {
    ctx = ctx[symbols.shadow] ?? ctx;
    if (["prototype", "then", "registry", "lifecycle"].includes(name)) return;
    if (name[0] === "$" || name[0] === "_") return;
    if (!ctx.runtime.plugin) return;
    if (ctx.bail(ctx, "internal/inject", name)) return;
    const lines = error.stack.split("\n");
    lines.splice(1, 1);
    error.stack = lines.join("\n");
    ctx.emit(ctx, "internal/warning", error);
  }
  get(name) {
    var _a15;
    const internal = this.ctx[symbols.internal][name];
    if ((internal == null ? void 0 : internal.type) !== "service") return;
    const key = this.ctx[symbols.isolate][name];
    const value = (_a15 = this.ctx[symbols.store][key]) == null ? void 0 : _a15.value;
    return getTraceable(this.ctx, value);
  }
  set(name, value) {
    var _a15;
    this.provide(name);
    const key = this.ctx[symbols.isolate][name];
    const oldValue = (_a15 = this.ctx[symbols.store][key]) == null ? void 0 : _a15.value;
    value ?? (value = void 0);
    let dispose = __name2(() => {
    }, "dispose");
    if (oldValue === value) return dispose;
    if (!isNullable(value) && !isNullable(oldValue)) {
      throw new Error(`service ${name} has been registered`);
    }
    const ctx = this.ctx;
    if (!isNullable(value)) {
      dispose = ctx.effect(() => () => {
        ctx.set(name, void 0);
      });
    }
    if (isUnproxyable(value)) {
      ctx.emit(ctx, "internal/warning", new Error(`service ${name} is an unproxyable object, which may lead to unexpected behavior`));
    }
    const self = Object.create(ctx);
    self[symbols.filter] = (ctx2) => {
      return ctx[symbols.isolate][name] === ctx2[symbols.isolate][name];
    };
    ctx.emit(self, "internal/before-service", name, value);
    ctx[symbols.store][key] = { value, source: ctx };
    ctx.emit(self, "internal/service", name, oldValue);
    return dispose;
  }
  provide(name, value, builtin) {
    const internal = this.ctx.root[symbols.internal];
    if (name in internal) return;
    const key = Symbol(name);
    internal[name] = { type: "service", builtin };
    this.ctx.root[symbols.isolate][name] = key;
    if (!isObject(value)) return;
    this.ctx[symbols.store][key] = { value, source: null };
    defineProperty(value, symbols.tracker, {
      associate: name,
      property: "ctx"
    });
  }
  _accessor(name, options) {
    const internal = this.ctx.root[symbols.internal];
    if (name in internal) return () => {
    };
    internal[name] = { type: "accessor", ...options };
    return () => delete this.ctx.root[symbols.isolate][name];
  }
  accessor(name, options) {
    this.ctx.scope.effect(() => {
      return this._accessor(name, options);
    });
  }
  alias(name, aliases) {
    const internal = this.ctx.root[symbols.internal];
    if (name in internal) return;
    for (const key of aliases) {
      internal[key] || (internal[key] = { type: "alias", name });
    }
  }
  _mixin(source, mixins) {
    const entries = Array.isArray(mixins) ? mixins.map((key) => [key, key]) : Object.entries(mixins);
    const getTarget = typeof source === "string" ? (ctx) => ctx[source] : () => source;
    const disposables = entries.map(([key, value]) => {
      return this._accessor(value, {
        get(receiver) {
          const service = getTarget(this);
          if (isNullable(service)) return service;
          const mixin = receiver ? withProps(receiver, service) : service;
          const value2 = Reflect.get(service, key, mixin);
          if (typeof value2 !== "function") return value2;
          return value2.bind(mixin ?? service);
        },
        set(value2, receiver) {
          const service = getTarget(this);
          const mixin = receiver ? withProps(receiver, service) : service;
          return Reflect.set(service, key, value2, mixin);
        }
      });
    });
    return () => disposables.forEach((dispose) => dispose());
  }
  mixin(source, mixins) {
    this.ctx.scope.effect(() => {
      return this._mixin(source, mixins);
    });
  }
  trace(value) {
    return getTraceable(this.ctx, value);
  }
  bind(callback) {
    return new Proxy(callback, {
      apply: __name2((target, thisArg, args) => {
        return target.apply(this.trace(thisArg), args.map((arg) => this.trace(arg)));
      }, "apply")
    });
  }
}, __name2(_a, "ReflectService"), __publicField(_a, "handler", {
  get: __name2((target, prop, ctx) => {
    if (typeof prop !== "string") return Reflect.get(target, prop, ctx);
    if (Reflect.has(target, prop)) {
      return getTraceable(ctx, Reflect.get(target, prop, ctx), true);
    }
    const [name, internal] = _a.resolveInject(target, prop);
    const error = new Error(`property ${name} is not registered, declare it as \`inject\` to suppress this warning`);
    if (!internal) {
      _a.checkInject(ctx, name, error);
      return Reflect.get(target, name, ctx);
    } else if (internal.type === "accessor") {
      return internal.get.call(ctx, ctx[symbols.receiver]);
    } else {
      if (!internal.builtin) _a.checkInject(ctx, name, error);
      return ctx.reflect.get(name);
    }
  }, "get"),
  set: __name2((target, prop, value, ctx) => {
    if (typeof prop !== "string") return Reflect.set(target, prop, value, ctx);
    const [name, internal] = _a.resolveInject(target, prop);
    if (!internal) {
      return Reflect.set(target, name, value, ctx);
    }
    if (internal.type === "accessor") {
      if (!internal.set) return false;
      return internal.set.call(ctx, value, ctx[symbols.receiver]);
    } else {
      ctx.reflect.set(name, value);
      return true;
    }
  }, "set"),
  has: __name2((target, prop) => {
    if (typeof prop !== "string") return Reflect.has(target, prop);
    if (Reflect.has(target, prop)) return true;
    const [, internal] = _a.resolveInject(target, prop);
    return !!internal;
  }, "has")
}), _a);
var reflect_default = ReflectService;
function isBailed(value) {
  return value !== null && value !== false && value !== void 0;
}
__name2(isBailed, "isBailed");
var _a2;
var Lifecycle = (_a2 = class {
  constructor(ctx) {
    __publicField(this, "isActive", false);
    __publicField(this, "_tasks", /* @__PURE__ */ new Set());
    __publicField(this, "_hooks", {});
    this.ctx = ctx;
    defineProperty(this, symbols.tracker, {
      associate: "lifecycle",
      property: "ctx"
    });
    defineProperty(this.on("internal/listener", function(name, listener, options) {
      const method = options.prepend ? "unshift" : "push";
      if (name === "ready") {
        if (!this.lifecycle.isActive) return;
        this.scope.ensure(async () => listener());
        return () => false;
      } else if (name === "dispose") {
        this.scope.disposables[method](listener);
        defineProperty(listener, "name", "event <dispose>");
        return () => remove(this.scope.disposables, listener);
      } else if (name === "fork") {
        this.scope.runtime.forkables[method](listener);
        return this.scope.collect("event <fork>", () => remove(this.scope.runtime.forkables, listener));
      }
    }), Context.static, ctx.scope);
    for (const level of ["info", "error", "warning"]) {
      defineProperty(this.on(`internal/${level}`, (format, ...param) => {
        if (this._hooks[`internal/${level}`].length > 1) return;
        console.info(format, ...param);
      }), Context.static, ctx.scope);
    }
    defineProperty(this.on("internal/before-service", function(name) {
      var _a15;
      for (const runtime of this.registry.values()) {
        if (!((_a15 = runtime.inject[name]) == null ? void 0 : _a15.required)) continue;
        const scopes = runtime.isReusable ? runtime.children : [runtime];
        for (const scope of scopes) {
          if (!this[symbols.filter](scope.ctx)) continue;
          scope.updateStatus();
          scope.reset();
        }
      }
    }, { global: true }), Context.static, ctx.scope);
    defineProperty(this.on("internal/service", function(name) {
      var _a15;
      for (const runtime of this.registry.values()) {
        if (!((_a15 = runtime.inject[name]) == null ? void 0 : _a15.required)) continue;
        const scopes = runtime.isReusable ? runtime.children : [runtime];
        for (const scope of scopes) {
          if (!this[symbols.filter](scope.ctx)) continue;
          scope.start();
        }
      }
    }, { global: true }), Context.static, ctx.scope);
    const checkInject = __name2((scope, name) => {
      if (!scope.runtime.plugin) return false;
      for (const key in scope.runtime.inject) {
        if (name === reflect_default.resolveInject(scope.ctx, key)[0]) return true;
      }
      return checkInject(scope.parent.scope, name);
    }, "checkInject");
    defineProperty(this.on("internal/inject", function(name) {
      return checkInject(this.scope, name);
    }, { global: true }), Context.static, ctx.scope);
  }
  async flush() {
    while (this._tasks.size) {
      await Promise.all(Array.from(this._tasks));
    }
  }
  filterHooks(hooks, thisArg) {
    thisArg = getTraceable(this.ctx, thisArg);
    return hooks.slice().filter((hook) => {
      const filter = thisArg == null ? void 0 : thisArg[Context.filter];
      return hook.global || !filter || filter.call(thisArg, hook.ctx);
    });
  }
  *dispatch(type, args) {
    const thisArg = typeof args[0] === "object" || typeof args[0] === "function" ? args.shift() : null;
    const name = args.shift();
    if (name !== "internal/event") {
      this.emit("internal/event", type, name, args, thisArg);
    }
    for (const hook of this.filterHooks(this._hooks[name] || [], thisArg)) {
      yield hook.callback.apply(thisArg, args);
    }
  }
  async parallel(...args) {
    await Promise.all(this.dispatch("emit", args));
  }
  emit(...args) {
    Array.from(this.dispatch("emit", args));
  }
  async serial(...args) {
    for await (const result of this.dispatch("serial", args)) {
      if (isBailed(result)) return result;
    }
  }
  bail(...args) {
    for (const result of this.dispatch("bail", args)) {
      if (isBailed(result)) return result;
    }
  }
  register(label, hooks, callback, options) {
    const method = options.prepend ? "unshift" : "push";
    hooks[method]({ ctx: this.ctx, callback, ...options });
    return this.ctx.state.collect(label, () => this.unregister(hooks, callback));
  }
  unregister(hooks, callback) {
    const index = hooks.findIndex((hook) => hook.callback === callback);
    if (index >= 0) {
      hooks.splice(index, 1);
      return true;
    }
  }
  on(name, listener, options) {
    var _a15;
    if (typeof options !== "object") {
      options = { prepend: options };
    }
    this.ctx.scope.assertActive();
    listener = this.ctx.reflect.bind(listener);
    const result = this.bail(this.ctx, "internal/listener", name, listener, options);
    if (result) return result;
    const hooks = (_a15 = this._hooks)[name] || (_a15[name] = []);
    const label = typeof name === "string" ? `event <${name}>` : "event (Symbol)";
    return this.register(label, hooks, listener, options);
  }
  once(name, listener, options) {
    const dispose = this.on(name, function(...args) {
      dispose();
      return listener.apply(this, args);
    }, options);
    return dispose;
  }
  async start() {
    this.isActive = true;
    const hooks = this._hooks.ready || [];
    while (hooks.length) {
      const { ctx, callback } = hooks.shift();
      ctx.scope.ensure(async () => callback());
    }
    await this.flush();
  }
  async stop() {
    this.isActive = false;
    this.ctx.scope.reset();
  }
}, __name2(_a2, "Lifecycle"), _a2);
var events_default = Lifecycle;
var ScopeStatus = ((ScopeStatus2) => {
  ScopeStatus2[ScopeStatus2["PENDING"] = 0] = "PENDING";
  ScopeStatus2[ScopeStatus2["LOADING"] = 1] = "LOADING";
  ScopeStatus2[ScopeStatus2["ACTIVE"] = 2] = "ACTIVE";
  ScopeStatus2[ScopeStatus2["FAILED"] = 3] = "FAILED";
  ScopeStatus2[ScopeStatus2["DISPOSED"] = 4] = "DISPOSED";
  return ScopeStatus2;
})(ScopeStatus || {});
var _a3;
var CordisError = (_a3 = class extends Error {
  constructor(code, message2) {
    super(message2 ?? _a3.Code[code]);
    this.code = code;
  }
}, __name2(_a3, "CordisError"), _a3);
((CordisError2) => {
  CordisError2.Code = {
    INACTIVE_EFFECT: "cannot create effect on inactive context"
  };
})(CordisError || (CordisError = {}));
var _a4;
var EffectScope = (_a4 = class {
  constructor(parent, config3) {
    __publicField(this, "uid");
    __publicField(this, "ctx");
    __publicField(this, "disposables", []);
    __publicField(this, "error");
    __publicField(this, "status", 0);
    __publicField(this, "isActive", false);
    // Same as `this.ctx`, but with a more specific type.
    __publicField(this, "context");
    __publicField(this, "proxy");
    __publicField(this, "acceptors", []);
    __publicField(this, "tasks", /* @__PURE__ */ new Set());
    __publicField(this, "hasError", false);
    this.parent = parent;
    this.config = config3;
    this.uid = parent.registry ? parent.registry.counter : 0;
    this.ctx = this.context = parent.extend({ scope: this });
    this.proxy = new Proxy({}, {
      get: __name2((target, key) => Reflect.get(this.config, key), "get")
    });
  }
  get _config() {
    return this.runtime.isReactive ? this.proxy : this.config;
  }
  assertActive() {
    if (this.uid !== null || this.isActive) return;
    throw new CordisError("INACTIVE_EFFECT");
  }
  effect(callback, config3) {
    this.assertActive();
    const result = isConstructor(callback) ? new callback(this.ctx, config3) : callback(this.ctx, config3);
    let disposed = false;
    const original2 = typeof result === "function" ? result : result.dispose.bind(result);
    const wrapped = __name2((...args) => {
      if (disposed) return;
      disposed = true;
      remove(this.disposables, wrapped);
      return original2(...args);
    }, "wrapped");
    this.disposables.push(wrapped);
    if (typeof result === "function") return wrapped;
    result.dispose = wrapped;
    return result;
  }
  collect(label, callback) {
    const dispose = defineProperty(() => {
      remove(this.disposables, dispose);
      return callback();
    }, "name", label);
    this.disposables.push(dispose);
    return dispose;
  }
  restart() {
    this.reset();
    this.error = null;
    this.hasError = false;
    this.status = 0;
    this.start();
  }
  _getStatus() {
    if (this.uid === null) return 4;
    if (this.hasError) return 3;
    if (this.tasks.size) return 1;
    if (this.ready) return 2;
    return 0;
  }
  updateStatus(callback) {
    const oldValue = this.status;
    callback == null ? void 0 : callback();
    this.status = this._getStatus();
    if (oldValue !== this.status) {
      this.context.emit("internal/status", this, oldValue);
    }
  }
  ensure(callback) {
    const task = callback().catch((reason) => {
      this.context.emit(this.ctx, "internal/error", reason);
      this.cancel(reason);
    }).finally(() => {
      this.updateStatus(() => this.tasks.delete(task));
      this.context.events._tasks.delete(task);
    });
    this.updateStatus(() => this.tasks.add(task));
    this.context.events._tasks.add(task);
  }
  cancel(reason) {
    this.error = reason;
    this.updateStatus(() => this.hasError = true);
    this.reset();
  }
  get ready() {
    return Object.entries(this.runtime.inject).every(([name, inject2]) => {
      return !inject2.required || !isNullable(this.ctx.get(name));
    });
  }
  reset() {
    this.isActive = false;
    this.disposables = this.disposables.splice(0).filter((dispose) => {
      if (this.uid !== null && dispose[Context.static] === this) return true;
      (async () => dispose())().catch((reason) => {
        this.context.emit(this.ctx, "internal/error", reason);
      });
    });
  }
  init(error) {
    if (!this.config) {
      this.cancel(error);
    } else {
      this.start();
    }
  }
  start() {
    if (!this.ready || this.isActive || this.uid === null) return true;
    this.isActive = true;
    this.updateStatus(() => this.hasError = false);
  }
  accept(...args) {
    const keys = Array.isArray(args[0]) ? args.shift() : null;
    const acceptor = { keys, callback: args[0], ...args[1] };
    return this.effect(() => {
      var _a15;
      this.acceptors.push(acceptor);
      if (acceptor.immediate) (_a15 = acceptor.callback) == null ? void 0 : _a15.call(acceptor, this.config);
      return () => remove(this.acceptors, acceptor);
    });
  }
  decline(keys) {
    return this.accept(keys, () => true);
  }
  checkUpdate(resolved2, forced) {
    if (forced || !this.config) return [true, true];
    if (forced === false) return [false, false];
    const modified = /* @__PURE__ */ Object.create(null);
    const checkPropertyUpdate = __name2((key) => {
      const result = modified[key] ?? (modified[key] = !deepEqual(this.config[key], resolved2[key]));
      hasUpdate || (hasUpdate = result);
      return result;
    }, "checkPropertyUpdate");
    const ignored = /* @__PURE__ */ new Set();
    let hasUpdate = false, shouldRestart = false;
    let fallback = this.runtime.isReactive || null;
    for (const { keys, callback, passive } of this.acceptors) {
      if (!keys) {
        fallback || (fallback = !passive);
      } else if (passive) {
        keys == null ? void 0 : keys.forEach((key) => ignored.add(key));
      } else {
        let hasUpdate2 = false;
        for (const key of keys) {
          hasUpdate2 || (hasUpdate2 = checkPropertyUpdate(key));
        }
        if (!hasUpdate2) continue;
      }
      const result = callback == null ? void 0 : callback(resolved2);
      if (result) shouldRestart = true;
    }
    for (const key in { ...this.config, ...resolved2 }) {
      if (fallback === false) continue;
      if (!(key in modified) && !ignored.has(key)) {
        const hasUpdate2 = checkPropertyUpdate(key);
        if (fallback === null) shouldRestart || (shouldRestart = hasUpdate2);
      }
    }
    return [hasUpdate, shouldRestart];
  }
}, __name2(_a4, "EffectScope"), _a4);
var _a5;
var ForkScope = (_a5 = class extends EffectScope {
  constructor(parent, runtime, config3, error) {
    super(parent, config3);
    __publicField(this, "dispose");
    this.runtime = runtime;
    this.dispose = defineProperty(parent.scope.collect(`fork <${parent.runtime.name}>`, () => {
      this.uid = null;
      this.reset();
      this.context.emit("internal/fork", this);
      const result = remove(runtime.disposables, this.dispose);
      if (remove(runtime.children, this) && !runtime.children.length) {
        parent.registry.delete(runtime.plugin);
      }
      return result;
    }), Context.static, runtime);
    runtime.children.push(this);
    runtime.disposables.push(this.dispose);
    this.context.emit("internal/fork", this);
    this.init(error);
  }
  start() {
    if (super.start()) return true;
    for (const fork of this.runtime.forkables) {
      this.ensure(async () => fork(this.context, this._config));
    }
  }
  update(config3, forced) {
    const oldConfig = this.config;
    const state = this.runtime.isForkable ? this : this.runtime;
    if (state.config !== oldConfig) return;
    let resolved2;
    try {
      resolved2 = resolveConfig(this.runtime.plugin, config3);
    } catch (error) {
      this.context.emit("internal/error", error);
      return this.cancel(error);
    }
    const [hasUpdate, shouldRestart] = state.checkUpdate(resolved2, forced);
    this.context.emit("internal/before-update", this, config3);
    this.config = resolved2;
    state.config = resolved2;
    if (hasUpdate) {
      this.context.emit("internal/update", this, oldConfig);
    }
    if (shouldRestart) state.restart();
  }
}, __name2(_a5, "ForkScope"), _a5);
var _a6;
var MainScope = (_a6 = class extends EffectScope {
  constructor(ctx, plugin, config3, error) {
    super(ctx, config3);
    __publicField(this, "value");
    __publicField(this, "runtime", this);
    __publicField(this, "schema");
    __publicField(this, "name");
    __publicField(this, "inject", /* @__PURE__ */ Object.create(null));
    __publicField(this, "forkables", []);
    __publicField(this, "children", []);
    __publicField(this, "isReusable", false);
    __publicField(this, "isReactive", false);
    __publicField(this, "apply", __name2((context, config3) => {
      if (typeof this.plugin !== "function") {
        return this.plugin.apply(context, config3);
      } else if (isConstructor(this.plugin)) {
        const instance = new this.plugin(context, config3);
        const name = instance[Context.expose];
        if (name) {
          context.set(name, instance);
        }
        if (instance["fork"]) {
          this.forkables.push(instance["fork"].bind(instance));
        }
        return instance;
      } else {
        return this.plugin(context, config3);
      }
    }, "apply"));
    this.plugin = plugin;
    if (!plugin) {
      this.name = "root";
      this.isActive = true;
    } else {
      this.setup();
      this.init(error);
    }
  }
  get isForkable() {
    return this.forkables.length > 0;
  }
  fork(parent, config3, error) {
    return new ForkScope(parent, this, config3, error);
  }
  dispose() {
    this.uid = null;
    this.reset();
    this.context.emit("internal/runtime", this);
    return true;
  }
  setup() {
    const { name } = this.plugin;
    if (name && name !== "apply") this.name = name;
    this.schema = this.plugin["Config"] || this.plugin["schema"];
    this.inject = Inject.resolve(this.plugin["using"] || this.plugin["inject"]);
    this.isReusable = this.plugin["reusable"];
    this.isReactive = this.plugin["reactive"];
    this.context.emit("internal/runtime", this);
    if (this.isReusable) {
      this.forkables.push(this.apply);
    }
  }
  reset() {
    super.reset();
    for (const fork of this.children) {
      fork.reset();
    }
  }
  start() {
    if (super.start()) return true;
    if (!this.isReusable && this.plugin) {
      this.ensure(async () => this.value = this.apply(this.ctx, this._config));
    }
    for (const fork of this.children) {
      fork.start();
    }
  }
  update(config3, forced) {
    if (this.isForkable) {
      const warning = new Error(`attempting to update forkable plugin "${this.plugin.name}", which may lead to unexpected behavior`);
      this.context.emit(this.ctx, "internal/warning", warning);
    }
    const oldConfig = this.config;
    let resolved2;
    try {
      resolved2 = resolveConfig(this.runtime.plugin || this.context.constructor, config3);
    } catch (error) {
      this.context.emit("internal/error", error);
      return this.cancel(error);
    }
    const [hasUpdate, shouldRestart] = this.checkUpdate(resolved2, forced);
    const state = this.children.find((fork) => fork.config === oldConfig);
    this.config = resolved2;
    if (state) {
      this.context.emit("internal/before-update", state, config3);
      state.config = resolved2;
      if (hasUpdate) {
        this.context.emit("internal/update", state, oldConfig);
      }
    }
    if (shouldRestart) this.restart();
  }
}, __name2(_a6, "MainScope"), _a6);
function isApplicable(object) {
  return object && typeof object === "object" && typeof object.apply === "function";
}
__name2(isApplicable, "isApplicable");
function Inject(inject2) {
  return function(value, ctx) {
    if (ctx.kind === "class") {
      value.inject = inject2;
    } else if (ctx.kind === "method") {
      ctx.addInitializer(function() {
        var _a15;
        const property = (_a15 = this[symbols.tracker]) == null ? void 0 : _a15.property;
        if (!property) throw new Error("missing context tracker");
        this[property].inject(inject2, (ctx2) => {
          value.call(withProps(this, { [property]: ctx2 }));
        });
      });
    } else {
      throw new Error("@Inject can only be used on class or class methods");
    }
  };
}
__name2(Inject, "Inject");
((Inject2) => {
  function resolve(inject2) {
    if (!inject2) return {};
    if (Array.isArray(inject2)) {
      return Object.fromEntries(inject2.map((name) => [name, { required: true }]));
    }
    const { required, optional, ...rest } = inject2;
    if (Array.isArray(required)) {
      Object.assign(rest, Object.fromEntries(required.map((name) => [name, { required: true }])));
    }
    if (Array.isArray(optional)) {
      Object.assign(rest, Object.fromEntries(optional.map((name) => [name, { required: false }])));
    }
    return rest;
  }
  Inject2.resolve = resolve;
  __name2(resolve, "resolve");
})(Inject || (Inject = {}));
var _a7;
var Registry = (_a7 = class {
  constructor(ctx, config3) {
    __publicField(this, "_counter", 0);
    __publicField(this, "_internal", /* @__PURE__ */ new Map());
    __publicField(this, "context");
    this.ctx = ctx;
    defineProperty(this, symbols.tracker, {
      associate: "registry",
      property: "ctx"
    });
    this.context = ctx;
    const runtime = new MainScope(ctx, null, config3);
    ctx.scope = runtime;
    runtime.ctx = ctx;
    this.set(null, runtime);
  }
  get counter() {
    return ++this._counter;
  }
  get size() {
    return this._internal.size;
  }
  resolve(plugin, assert = false) {
    if (plugin === null) return plugin;
    if (typeof plugin === "function") return plugin;
    if (isApplicable(plugin)) return plugin.apply;
    if (assert) throw new Error('invalid plugin, expect function or object with an "apply" method, received ' + typeof plugin);
  }
  get(plugin) {
    const key = this.resolve(plugin);
    return key && this._internal.get(key);
  }
  has(plugin) {
    const key = this.resolve(plugin);
    return !!key && this._internal.has(key);
  }
  set(plugin, state) {
    const key = this.resolve(plugin);
    this._internal.set(key, state);
  }
  delete(plugin) {
    const key = this.resolve(plugin);
    const runtime = key && this._internal.get(key);
    if (!runtime) return;
    this._internal.delete(key);
    runtime.dispose();
    return runtime;
  }
  keys() {
    return this._internal.keys();
  }
  values() {
    return this._internal.values();
  }
  entries() {
    return this._internal.entries();
  }
  forEach(callback) {
    return this._internal.forEach(callback);
  }
  using(inject2, callback) {
    return this.inject(inject2, callback);
  }
  inject(inject2, callback) {
    return this.plugin({ inject: inject2, apply: callback, name: callback.name });
  }
  plugin(plugin, config3, error) {
    this.resolve(plugin, true);
    this.ctx.scope.assertActive();
    if (!error) {
      try {
        config3 = resolveConfig(plugin, config3);
      } catch (reason) {
        this.context.emit(this.ctx, "internal/error", reason);
        error = reason;
        config3 = null;
      }
    }
    let runtime = this.get(plugin);
    if (runtime) {
      if (!runtime.isForkable) {
        this.context.emit(this.ctx, "internal/warning", new Error(`duplicate plugin detected: ${plugin.name}`));
      }
      return runtime.fork(this.ctx, config3, error);
    }
    runtime = new MainScope(this.ctx, plugin, config3, error);
    this.set(plugin, runtime);
    return runtime.fork(this.ctx, config3, error);
  }
}, __name2(_a7, "Registry"), _a7);
var registry_default = Registry;
var _a8;
var Context = (_a8 = class {
  static is(value) {
    return !!(value == null ? void 0 : value[_a8.is]);
  }
  /** @deprecated use `Service.traceable` instead */
  static associate(object, name) {
    return object;
  }
  constructor(config3) {
    config3 = resolveConfig(this.constructor, config3);
    this[symbols.store] = /* @__PURE__ */ Object.create(null);
    this[symbols.isolate] = /* @__PURE__ */ Object.create(null);
    this[symbols.internal] = /* @__PURE__ */ Object.create(null);
    this[symbols.intercept] = /* @__PURE__ */ Object.create(null);
    const self = new Proxy(this, reflect_default.handler);
    self.root = self;
    self.reflect = new reflect_default(self);
    self.registry = new registry_default(self, config3);
    self.lifecycle = new events_default(self);
    const attach = __name2((internal) => {
      var _a15;
      if (!internal) return;
      attach(Object.getPrototypeOf(internal));
      for (const key of Object.getOwnPropertyNames(internal)) {
        const constructor = (_a15 = internal[key]["prototype"]) == null ? void 0 : _a15.constructor;
        if (!constructor) continue;
        self[internal[key]["key"]] = new constructor(self, config3);
        defineProperty(self[internal[key]["key"]], "ctx", self);
      }
    }, "attach");
    attach(this[symbols.internal]);
    return self;
  }
  [Symbol.for("nodejs.util.inspect.custom")]() {
    return `Context <${this.name}>`;
  }
  get name() {
    let runtime = this.runtime;
    while (runtime && !runtime.name) {
      runtime = runtime.parent.runtime;
    }
    return runtime == null ? void 0 : runtime.name;
  }
  get events() {
    return this.lifecycle;
  }
  /** @deprecated */
  get state() {
    return this.scope;
  }
  extend(meta = {}) {
    var _a15;
    const source = (_a15 = Reflect.getOwnPropertyDescriptor(this, symbols.shadow)) == null ? void 0 : _a15.value;
    const self = Object.assign(Object.create(getTraceable(this, this)), meta);
    if (!source) return self;
    return Object.assign(Object.create(self), { [symbols.shadow]: source });
  }
  isolate(name, label) {
    const shadow = Object.create(this[symbols.isolate]);
    shadow[name] = label ?? Symbol(name);
    return this.extend({ [symbols.isolate]: shadow });
  }
  intercept(name, config3) {
    const intercept = Object.create(this[symbols.intercept]);
    intercept[name] = config3;
    return this.extend({ [symbols.intercept]: intercept });
  }
}, __name2(_a8, "Context"), __publicField(_a8, "store", symbols.store), __publicField(_a8, "events", symbols.events), __publicField(_a8, "static", symbols.static), __publicField(_a8, "filter", symbols.filter), __publicField(_a8, "expose", symbols.expose), __publicField(_a8, "isolate", symbols.isolate), __publicField(_a8, "internal", symbols.internal), __publicField(_a8, "intercept", symbols.intercept), __publicField(_a8, "origin", "ctx"), __publicField(_a8, "current", "ctx"), _a8.is[Symbol.toPrimitive] = () => Symbol.for("cordis.is"), _a8.prototype[_a8.is] = true, _a8);
Context.prototype[Context.internal] = /* @__PURE__ */ Object.create(null);
var _a9;
var Service = (_a9 = class {
  constructor(...args) {
    __publicField(this, "ctx");
    __publicField(this, "name");
    __publicField(this, "config");
    let _ctx, name, immediate, config3;
    if (Context.is(args[0])) {
      _ctx = args[0];
      if (typeof args[1] === "string") {
        name = args[1];
        immediate = args[2];
      } else {
        config3 = args[1];
      }
    } else {
      config3 = args[0];
    }
    name ?? (name = this.constructor[symbols.provide]);
    immediate ?? (immediate = this.constructor[symbols.immediate]);
    let self = this;
    const tracker = {
      associate: name,
      property: "ctx"
    };
    if (self[symbols.invoke]) {
      self = createCallable(name, joinPrototype(Object.getPrototypeOf(this), Function.prototype), tracker);
    }
    if (_ctx) {
      self.ctx = _ctx;
    } else {
      self[symbols.setup]();
    }
    self.name = name;
    self.config = config3;
    defineProperty(self, symbols.tracker, tracker);
    self.ctx.provide(name);
    self.ctx.runtime.name = name;
    if (immediate) {
      if (_ctx) self[symbols.expose] = name;
      else self.ctx.set(name, self);
    }
    self.ctx.on("ready", async () => {
      await Promise.resolve();
      await self.start();
      if (!immediate) self.ctx.set(name, self);
    });
    self.ctx.on("dispose", () => self.stop());
    return self;
  }
  start() {
  }
  stop() {
  }
  [symbols.filter](ctx) {
    return ctx[symbols.isolate][this.name] === this.ctx[symbols.isolate][this.name];
  }
  [symbols.setup]() {
    this.ctx = new Context();
  }
  [symbols.extend](props) {
    let self;
    if (this[_a9.invoke]) {
      self = createCallable(this.name, this, this[symbols.tracker]);
    } else {
      self = Object.create(this);
    }
    return Object.assign(self, props);
  }
  static [Symbol.hasInstance](instance) {
    var _a15;
    let constructor = instance.constructor;
    while (constructor) {
      constructor = (_a15 = constructor.prototype) == null ? void 0 : _a15.constructor;
      if (constructor === this) return true;
      constructor = Object.getPrototypeOf(constructor);
    }
    return false;
  }
}, __name2(_a9, "Service"), __publicField(_a9, "setup", symbols.setup), __publicField(_a9, "invoke", symbols.invoke), __publicField(_a9, "extend", symbols.extend), __publicField(_a9, "tracker", symbols.tracker), __publicField(_a9, "provide", symbols.provide), __publicField(_a9, "immediate", symbols.immediate), _a9);

// node_modules/.bun/reggol@1.7.1/node_modules/reggol/lib/browser.mjs
var import_supports_color = __toESM(require_browser(), 1);
var import_object_inspect = __toESM(require_object_inspect(), 1);
var __create = Object.create;
var __defProp3 = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __name3 = (target, value) => __defProp3(target, "name", { value, configurable: true });
var __commonJS2 = (cb, mod) => function __require() {
  return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp3(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toESM2 = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
  // If the importer is in node compatibility mode or this is not an ESM
  // file that has been converted to a CommonJS file using a Babel-
  // compatible transform (i.e. "__esModule" has not been set), then set
  // "default" to the CommonJS "module.exports" for node compatibility.
  isNodeMode || !mod || !mod.__esModule ? __defProp3(target, "default", { value: mod, enumerable: true }) : target,
  mod
));
var require_shared = __commonJS2({
  "src/shared.ts"(exports, module) {
    "use strict";
    var _a15;
    var c16 = [6, 2, 3, 4, 5, 1];
    var c256 = [
      20,
      21,
      26,
      27,
      32,
      33,
      38,
      39,
      40,
      41,
      42,
      43,
      44,
      45,
      56,
      57,
      62,
      63,
      68,
      69,
      74,
      75,
      76,
      77,
      78,
      79,
      80,
      81,
      92,
      93,
      98,
      99,
      112,
      113,
      129,
      134,
      135,
      148,
      149,
      160,
      161,
      162,
      163,
      164,
      165,
      166,
      167,
      168,
      169,
      170,
      171,
      172,
      173,
      178,
      179,
      184,
      185,
      196,
      197,
      198,
      199,
      200,
      201,
      202,
      203,
      204,
      205,
      206,
      207,
      208,
      209,
      214,
      215,
      220,
      221
    ];
    function isAggregateError(error) {
      return error instanceof Error && Array.isArray(error["errors"]);
    }
    __name3(isAggregateError, "isAggregateError");
    var Logger = (_a15 = class {
      constructor(name, meta) {
        __publicField(this, "extend", __name3((namespace) => {
          return new _a15(`${this.name}:${namespace}`, this.meta);
        }, "extend"));
        __publicField(this, "warning", __name3((format, ...args) => {
          this.warn(format, ...args);
        }, "warning"));
        this.name = name;
        this.meta = meta;
        this.createMethod("success", _a15.SUCCESS);
        this.createMethod("error", _a15.ERROR);
        this.createMethod("info", _a15.INFO);
        this.createMethod("warn", _a15.WARN);
        this.createMethod("debug", _a15.DEBUG);
      }
      static format(name, formatter) {
        this.formatters[name] = formatter;
      }
      static color(target, code, value, decoration = "") {
        if (!target.colors) return "" + value;
        return `\x1B[3${code < 8 ? code : "8;5;" + code}${target.colors >= 2 ? decoration : ""}m${value}\x1B[0m`;
      }
      static code(name, target) {
        let hash = 0;
        for (let i = 0; i < name.length; i++) {
          hash = (hash << 3) - hash + name.charCodeAt(i) + 13;
          hash |= 0;
        }
        const colors = !target.colors ? [] : target.colors >= 2 ? c256 : c16;
        return colors[Math.abs(hash) % colors.length];
      }
      static render(target, record) {
        var _a16, _b, _c;
        const prefix = `[${record.type[0].toUpperCase()}]`;
        const space = " ".repeat(((_a16 = target.label) == null ? void 0 : _a16.margin) ?? 1);
        let indent = 3 + space.length, output = "";
        if (target.showTime) {
          indent += target.showTime.length + space.length;
          output += _a15.color(target, 8, Time.template(target.showTime)) + space;
        }
        const code = _a15.code(record.name, target);
        const label = _a15.color(target, code, record.name, ";1");
        const padLength = (((_b = target.label) == null ? void 0 : _b.width) ?? 0) + label.length - record.name.length;
        if (((_c = target.label) == null ? void 0 : _c.align) === "right") {
          output += label.padStart(padLength) + space + prefix + space;
          indent += (target.label.width ?? 0) + space.length;
        } else {
          output += prefix + space + label.padEnd(padLength) + space;
        }
        output += record.content.replace(/\n/g, "\n" + " ".repeat(indent));
        if (target.showDiff && target.timestamp) {
          const diff = record.timestamp - target.timestamp;
          output += _a15.color(target, code, " +" + Time.format(diff));
        }
        return output;
      }
      createMethod(type, level) {
        this[type] = (...args) => {
          if (args.length === 1 && args[0] instanceof Error) {
            if (args[0].cause) {
              this[type](args[0].cause);
            } else if (isAggregateError(args[0])) {
              args[0].errors.forEach((error) => this[type](error));
              return;
            }
          }
          const id = ++_a15.id;
          const timestamp = Date.now();
          for (const target of _a15.targets) {
            if (this.getLevel(target) < level) continue;
            const content = this.format(target, ...args);
            const record = { id, type, level, name: this.name, meta: this.meta, content, timestamp };
            if (target.record) {
              target.record(record);
            } else {
              const { print = console.log } = target;
              print(_a15.render(target, record));
            }
            target.timestamp = timestamp;
          }
        };
      }
      format(target, ...args) {
        if (args[0] instanceof Error) {
          args[0] = args[0].stack || args[0].message;
          args.unshift("%s");
        } else if (typeof args[0] !== "string") {
          args.unshift("%o");
        }
        let format = args.shift();
        format = format.replace(/%([a-zA-Z%])/g, (match, char) => {
          if (match === "%%") return "%";
          const formatter = _a15.formatters[char];
          if (typeof formatter === "function") {
            const value = args.shift();
            return formatter(value, target, this);
          }
          return match;
        });
        for (let arg of args) {
          if (typeof arg === "object" && arg) {
            arg = _a15.formatters["o"](arg, target, this);
          }
          format += " " + arg;
        }
        const { maxLength = 10240 } = target;
        return format.split(/\r?\n/g).map((line) => {
          return line.slice(0, maxLength) + (line.length > maxLength ? "..." : "");
        }).join("\n");
      }
      getLevel(target) {
        const paths = this.name.split(":");
        let config3 = (target == null ? void 0 : target.levels) || _a15.levels;
        do {
          config3 = config3[paths.shift()] ?? config3["base"];
        } while (paths.length && typeof config3 === "object");
        return config3;
      }
      get level() {
        return this.getLevel();
      }
      set level(value) {
        const paths = this.name.split(":");
        let config3 = _a15.levels;
        while (paths.length > 1) {
          const name = paths.shift();
          const value2 = config3[name];
          if (typeof value2 === "object") {
            config3 = value2;
          } else {
            config3 = config3[name] = { base: value2 ?? config3.base };
          }
        }
        config3[paths[0]] = value;
      }
    }, __name3(_a15, "Logger"), // log levels
    __publicField(_a15, "SILENT", 0), __publicField(_a15, "SUCCESS", 1), __publicField(_a15, "ERROR", 1), __publicField(_a15, "INFO", 2), __publicField(_a15, "WARN", 2), __publicField(_a15, "DEBUG", 3), // global config
    __publicField(_a15, "id", 0), __publicField(_a15, "targets", [{
      colors: import_supports_color.stdout && import_supports_color.stdout.level,
      print(text) {
        console.log(text);
      }
    }]), // global registry
    __publicField(_a15, "formatters", /* @__PURE__ */ Object.create(null)), __publicField(_a15, "levels", {
      base: 2
    }), _a15);
    Logger.format("s", (value) => value);
    Logger.format("d", (value) => +value);
    Logger.format("j", (value) => JSON.stringify(value));
    Logger.format("c", (value, target, logger) => {
      return Logger.color(target, Logger.code(logger.name, target), value);
    });
    Logger.format("C", (value, target) => {
      return Logger.color(target, 15, value, ";1");
    });
    module.exports = Logger;
  }
});
var require_browser2 = __commonJS2({
  "src/browser.ts"(exports, module) {
    var import_shared = __toESM2(require_shared());
    import_shared.default.format("o", (value, target) => {
      return (0, import_object_inspect.default)(value, { depth: Infinity }).replace(/\s*\n\s*/g, " ");
    });
    module.exports = import_shared.default;
  }
});
var browser_default = require_browser2();

// node_modules/.bun/@cordisjs+logger@0.3.3+aa774b886391d1bc/node_modules/@cordisjs/logger/lib/index.mjs
var __defProp4 = Object.defineProperty;
var __name4 = (target, value) => __defProp4(target, "name", { value, configurable: true });
var _a10;
var LoggerService = (_a10 = class extends Service {
  constructor(ctx) {
    super(ctx, "logger", true);
    ctx.on("internal/info", function(format, ...args) {
      this.logger("app").info(format, ...args);
    });
    ctx.on("internal/error", function(format, ...args) {
      this.logger("app").error(format, ...args);
    });
    ctx.on("internal/warning", function(format, ...args) {
      this.logger("app").warn(format, ...args);
    });
  }
  [Service.invoke](name) {
    return new browser_default(name, defineProperty({}, "ctx", this.ctx));
  }
}, __name4(_a10, "LoggerService"), (() => {
  for (const type of ["success", "error", "info", "warn", "debug", "extend"]) {
    _a10.prototype[type] = function(...args) {
      return this(this.ctx.name)[type](...args);
    };
  }
})(), _a10);

// node_modules/.bun/@cordisjs+schema@0.1.1+aa774b886391d1bc/node_modules/@cordisjs/schema/lib/index.mjs
var __defProp5 = Object.defineProperty;
var __name5 = (target, value) => __defProp5(target, "name", { value, configurable: true });
var kSchemaOrder = Symbol("cordis.schema.order");
var _a11;
var SchemaService = (_a11 = class {
  constructor(ctx) {
    __publicField(this, "_data", lib_default.intersect([]));
    this.ctx = ctx;
    defineProperty(this, Service.tracker, {
      property: "ctx"
    });
  }
  extend(schema, order = 0) {
    const index = this._data.list.findIndex((a) => a[kSchemaOrder] < order);
    schema[kSchemaOrder] = order;
    return this.ctx.effect(() => {
      if (index >= 0) {
        this._data.list.splice(index, 0, schema);
      } else {
        this._data.list.push(schema);
      }
      this.ctx.emit("internal/service-schema");
      return () => {
        remove(this._data.list, schema);
        this.ctx.emit("internal/service-schema");
      };
    });
  }
  toJSON() {
    return this._data.toJSON();
  }
}, __name5(_a11, "SchemaService"), _a11);

// node_modules/.bun/@cordisjs+timer@0.3.2+aa774b886391d1bc/node_modules/@cordisjs/timer/lib/index.mjs
var __defProp6 = Object.defineProperty;
var __name6 = (target, value) => __defProp6(target, "name", { value, configurable: true });
var _a12;
var TimerService = (_a12 = class extends Service {
  constructor(ctx) {
    super(ctx, "timer", true);
    ctx.mixin("timer", ["setTimeout", "setInterval", "sleep", "throttle", "debounce"]);
  }
  setTimeout(callback, delay) {
    const dispose = this[Context.current].effect(() => {
      const timer = setTimeout(() => {
        dispose();
        callback();
      }, delay);
      return () => clearTimeout(timer);
    });
    return dispose;
  }
  setInterval(callback, delay) {
    return this[Context.current].effect(() => {
      const timer = setInterval(callback, delay);
      return () => clearInterval(timer);
    });
  }
  sleep(delay) {
    const caller = this[Context.current];
    return new Promise((resolve, reject) => {
      const dispose1 = this.setTimeout(() => {
        dispose1();
        dispose2();
        resolve();
      }, delay);
      const dispose2 = caller.on("dispose", () => {
        dispose1();
        dispose2();
        reject(new Error("Context has been disposed"));
      });
    });
  }
  createWrapper(callback, isDisposed = false) {
    const caller = this[Context.current];
    caller.scope.assertActive();
    let timer;
    const dispose = __name6(() => {
      isDisposed = true;
      remove(caller.scope.disposables, dispose);
      clearTimeout(timer);
    }, "dispose");
    const wrapper = __name6((...args) => {
      clearTimeout(timer);
      timer = callback(args, () => !isDisposed && caller.scope.isActive);
    }, "wrapper");
    wrapper.dispose = dispose;
    caller.scope.disposables.push(dispose);
    return wrapper;
  }
  throttle(callback, delay, noTrailing) {
    let lastCall = -Infinity;
    const execute = __name6((...args) => {
      lastCall = Date.now();
      callback(...args);
    }, "execute");
    return this.createWrapper((args, isActive) => {
      const now = Date.now();
      const remaining = delay - (now - lastCall);
      if (remaining <= 0) {
        execute(...args);
      } else if (isActive()) {
        return setTimeout(execute, remaining, ...args);
      }
    }, noTrailing);
  }
  debounce(callback, delay) {
    return this.createWrapper((args, isActive) => {
      if (!isActive())
        return;
      return setTimeout(callback, delay, ...args);
    });
  }
}, __name6(_a12, "TimerService"), _a12);

// node_modules/.bun/cordis@3.18.1+66361b782ecf3379/node_modules/cordis/lib/index.mjs
var __defProp7 = Object.defineProperty;
var __name7 = (target, value) => __defProp7(target, "name", { value, configurable: true });
var _a13;
var Context2 = (_a13 = class extends Context {
  constructor(config3) {
    var _a15, _b;
    super(config3);
    __publicField(this, "baseDir");
    this.baseDir = ((_b = (_a15 = globalThis.process) == null ? void 0 : _a15.cwd) == null ? void 0 : _b.call(_a15)) || "";
    this.provide("logger", void 0, true);
    this.provide("timer", void 0, true);
    this.plugin(LoggerService);
    this.plugin(TimerService);
  }
}, __name7(_a13, "Context"), _a13);
var _a14;
var Service2 = (_a14 = class extends Service {
  constructor(...args) {
    super(...args);
    /** @deprecated use `this.ctx.logger` instead */
    __publicField(this, "logger");
    __publicField(this, "schema");
    this.logger = this.ctx.logger(this.name);
    this.schema = new SchemaService(this.ctx);
  }
  [Service.setup]() {
    this.ctx = new Context2();
  }
}, __name7(_a14, "Service"), _a14);
function src_default3() {
}
__name7(src_default3, "default");

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/utils.ts
var Service3 = class extends Service2 {
};
function insert(list, item) {
  markRaw(item);
  const index = list.findIndex((a) => a.order < item.order);
  if (index >= 0) {
    list.splice(index, 0, item);
  } else {
    list.push(item);
  }
}

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/plugins/action.ts
function useMenu(id) {
  const ctx = useContext();
  return (event, value) => {
    ctx.define(id, value);
    event.preventDefault();
    const { clientX, clientY } = event;
    ctx.internal.activeMenus.splice(0, Infinity, {
      id,
      relative: {
        left: clientX,
        top: clientY,
        right: clientX,
        bottom: clientY
      }
    });
  };
}
var ActionService = class extends Service3 {
  constructor(ctx) {
    super(ctx, "$action", true);
    ctx.mixin("$action", ["action", "menu", "define"]);
    ctx.internal.scope = shallowReactive({});
    ctx.internal.menus = reactive({});
    ctx.internal.actions = reactive({});
    ctx.internal.activeMenus = reactive([]);
    ctx.addEventListener("keydown", (event) => {
      var _a15;
      const scope = this.createScope();
      for (const action of Object.values(ctx.internal.actions)) {
        if (!action.shortcut) continue;
        const keys = action.shortcut.split("+").map((key) => key.toLowerCase().trim());
        let ctrlKey = false, shiftKey = false, metaKey = false, code;
        for (const key of keys) {
          switch (key) {
            case "shift":
              shiftKey = true;
              continue;
            case "ctrl":
              if (navigator.platform.toLowerCase().includes("mac")) {
                metaKey = true;
              } else {
                ctrlKey = true;
              }
              continue;
            default:
              code = key;
          }
        }
        if (ctrlKey !== event.ctrlKey) continue;
        if (shiftKey !== event.shiftKey) continue;
        if (metaKey !== event.metaKey) continue;
        if (code !== event.key.toLowerCase()) continue;
        if ((_a15 = action.disabled) == null ? void 0 : _a15.call(action, scope)) continue;
        event.preventDefault();
        action.action(scope);
      }
    });
  }
  action(id, options) {
    markRaw(options);
    return this.ctx.effect(() => {
      this.ctx.internal.actions[id] = options;
      return () => delete this.ctx.internal.actions[id];
    });
  }
  menu(id, items) {
    return this.ctx.effect(() => {
      var _a15;
      const list = (_a15 = this.ctx.internal.menus)[id] || (_a15[id] = []);
      items.forEach((item) => insert(list, item));
      return () => {
        items.forEach((item) => remove(list, item));
        if (!list.length) delete this.ctx.internal.menus[id];
      };
    });
  }
  define(key, value) {
    return this.ctx.effect(() => {
      this.ctx.internal.scope[key] = value;
      return () => delete this.ctx.internal.scope[key];
    });
  }
  createScope(override = {}) {
    const scope = { ...this.ctx.internal.scope, ...override };
    return createScope(scope);
  }
};
function createScope(scope, prefix = "") {
  return new Proxy({}, {
    get: (target, key) => {
      if (typeof key === "symbol") return target[key];
      key = prefix + key;
      if (key in scope) return toValue(scope[key]);
      const _prefix = key + ".";
      if (Object.keys(scope).some((k) => k.startsWith(_prefix))) {
        return createScope(scope, key + ".");
      }
    }
  });
}

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/plugins/setting.ts
var useStorage = (key, version, fallback) => {
  const initial = fallback ? fallback() : {};
  initial["__version__"] = version;
  const storage = useLocalStorage("koishi.console." + key, initial);
  if (storage.value["__version__"] !== version) {
    storage.value = initial;
  }
  return storage;
};
function provideStorage(factory) {
  useStorage = factory;
}
function createStorage(key, version, fallback) {
  const storage = useLocalStorage("koishi.console." + key, {});
  const initial = fallback ? fallback() : {};
  if (storage.value.version !== version) {
    storage.value = { version, data: initial };
  } else if (!Array.isArray(storage.value.data)) {
    storage.value.data = { ...initial, ...storage.value.data };
  }
  return reactive(storage.value["data"]);
}
var original = useStorage("config", void 0, () => ({
  theme: {
    mode: "auto",
    dark: "default-dark",
    light: "default-light"
  },
  locale: "zh-CN"
}));
var resolved = ref({});
var useConfig = (useOriginal = false) => useOriginal ? original : resolved;
var SettingService = class extends Service3 {
  constructor(ctx) {
    super(ctx, "$setting", true);
    ctx.mixin("$setting", {
      settings: "settings",
      extendSchema: "schema"
    });
    ctx.internal.settings = reactive({});
    this.settings({
      id: "",
      title: "通用设置",
      order: 1e3,
      schema: lib_default.object({
        locale: lib_default.union(["zh-CN", "en-US"]).description("语言设置。")
      }).description("通用设置")
    });
    const schema = computed(() => {
      const list = [];
      for (const settings of Object.values(ctx.internal.settings)) {
        for (const options of settings) {
          if (options.schema) {
            list.push(options.schema);
          }
        }
      }
      return lib_default.intersect(list);
    });
    const doWatch = () => watch(resolved, (value) => {
      console.debug("config", JSON.stringify(value));
      original.value = schema.value.simplify(value);
    }, { deep: true });
    let stop = doWatch();
    const update = () => {
      stop == null ? void 0 : stop();
      try {
        resolved.value = schema.value(original.value);
      } catch (error) {
        console.error(error);
      }
      stop = doWatch();
    };
    ctx.effect(() => () => stop == null ? void 0 : stop());
    ctx.effect(() => watch(original, update, { deep: true }));
    ctx.effect(() => watch(schema, update));
  }
  extendSchema(extension) {
    extension.component = this.ctx.wrapComponent(extension.component);
    return this.ctx.effect(() => {
      src_default.extensions.add(extension);
      return () => src_default.extensions.delete(extension);
    });
  }
  settings(options) {
    markRaw(options);
    options.order ?? (options.order = 0);
    options.component = this.ctx.wrapComponent(options.component);
    return this.ctx.effect(() => {
      var _a15, _b;
      const list = (_a15 = this.ctx.internal.settings)[_b = options.id] || (_a15[_b] = []);
      insert(list, options);
      return () => {
        remove(list, options);
        if (!list.length) delete this.ctx.internal.settings[options.id];
      };
    });
  }
};

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/plugins/i18n.ts
var config = useConfig();
var I18nService = class extends Service3 {
  constructor(ctx) {
    super(ctx, "$i18n", true);
    __publicField(this, "i18n", createI18n({
      legacy: false,
      fallbackLocale: "zh-CN"
    }));
    ctx.effect(() => watchEffect(() => {
      this.i18n.global.locale.value = config.value.locale;
    }, { flush: "post" }));
  }
};

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/plugins/loader.ts
function defineExtension(callback) {
  return callback;
}
function unwrapExports(module) {
  return (module == null ? void 0 : module.default) || module;
}
var loaders = {
  async [`.css`](ctx, url) {
    const link = document.createElement("link");
    link.rel = "stylesheet";
    link.href = url;
    await new Promise((resolve, reject) => {
      link.onload = resolve;
      link.onerror = reject;
      ctx.effect(() => {
        document.head.appendChild(link);
        return () => document.head.removeChild(link);
      });
    });
  },
  async [``](ctx, url) {
    const exports = await import(
      /* @vite-ignore */
      url
    );
    ctx.plugin(unwrapExports(exports), ctx.extension.data);
  }
};
var LoaderService = class extends Service3 {
  constructor(ctx) {
    super(ctx, "$loader", true);
    __publicField(this, "backendId");
    __publicField(this, "extensions", shallowReactive({}));
    __publicField(this, "initTask", new Promise((resolve) => {
      watch(() => store.entry, async (newValue, oldValue) => {
        const { _id, ...rest } = newValue || {};
        if (this.backendId && _id && this.backendId !== _id) {
          window.location.reload();
          return;
        }
        this.backendId = _id;
        for (const key in this.extensions) {
          if (rest[key]) continue;
          this.extensions[key].scope.dispose();
          delete this.extensions[key];
        }
        await Promise.all(Object.entries(rest).map(([key, { files, paths, data }]) => {
          if (this.extensions[key]) return;
          const scope = this.ctx.isolate("extension").plugin(() => {
          });
          scope.ctx.extension = this.extensions[key] = { done: ref(false), scope, paths, data: ref(data) };
          const task = Promise.all(files.map((url) => {
            for (const ext in loaders) {
              if (!url.endsWith(ext)) continue;
              return loaders[ext](scope.ctx, url);
            }
          }));
          task.finally(() => this.extensions[key].done.value = true);
        }));
        if (!oldValue) resolve();
      }, { deep: true });
    }));
    receive("entry-data", ({ id, data }) => {
      var _a15;
      const entry = (_a15 = store.entry) == null ? void 0 : _a15[id];
      if (!entry) return;
      entry.data = data;
      this.extensions[id].data.value = data;
    });
  }
};

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/plugins/router.ts
import Overlay from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/chat/overlay.vue";
function getActivityId(path) {
  return path.split("/").find(Boolean) ?? "";
}
var redirectTo = ref();
var Activity = class {
  constructor(ctx, options) {
    this.ctx = ctx;
    this.options = options;
    __publicField(this, "id");
    __publicField(this, "_disposables", []);
    options.order ?? (options.order = 0);
    options.position ?? (options.position = "top");
    Object.assign(this, omit(options, ["icon", "name", "desc", "disabled"]));
    const { path, id = getActivityId(path), component } = options;
    this._disposables.push(ctx.$router.router.addRoute({ path, name: id, component, meta: { activity: this } }));
    this.id ?? (this.id = id);
    this.handleUpdate();
    this.authority ?? (this.authority = 0);
    this.fields ?? (this.fields = []);
    ctx.$router.pages[this.id] = this;
  }
  handleUpdate() {
    if (redirectTo.value) {
      const location2 = this.ctx.$router.router.resolve(redirectTo.value);
      if (location2.matched.length) {
        redirectTo.value = null;
        this.ctx.$router.router.replace(location2);
      }
    }
  }
  get icon() {
    return toValue(this.options.icon ?? "activity:default");
  }
  get name() {
    return toValue(this.options.name ?? this.id);
  }
  get desc() {
    return toValue(this.options.desc);
  }
  disabled() {
    var _a15, _b;
    if (this.ctx.bail("activity", this)) return true;
    if (!this.fields.every((key) => store[key])) return true;
    if (this.when && !this.when()) return true;
    if ((_b = (_a15 = this.options).disabled) == null ? void 0 : _b.call(_a15)) return true;
  }
  dispose() {
    var _a15;
    this._disposables.forEach((dispose) => dispose());
    const current = this.ctx.$router.router.currentRoute.value;
    if (((_a15 = current == null ? void 0 : current.meta) == null ? void 0 : _a15.activity) === this) {
      redirectTo.value = current.fullPath;
      this.ctx.$router.router.push(this.ctx.$router.cache["home"] || "/");
    }
    return delete this.ctx.$router.pages[this.id];
  }
};
var RouterService = class extends Service3 {
  constructor(ctx) {
    super(ctx, "$router", true);
    __publicField(this, "views", reactive({}));
    __publicField(this, "cache", reactive({}));
    __publicField(this, "pages", reactive({}));
    __publicField(this, "router", createRouter({
      history: createWebHistory(global2.uiPath),
      linkActiveClass: "active",
      routes: []
    }));
    ctx.mixin("$router", ["slot", "page"]);
    const initialTitle = document.title;
    ctx.effect(() => this.router.afterEach((route) => {
      const { name, fullPath } = this.router.currentRoute.value;
      this.cache[name] = fullPath;
      if (route.meta.activity) {
        document.title = `${route.meta.activity.name}`;
        if (initialTitle) document.title += ` | ${initialTitle}`;
      }
    }));
    this.router.beforeEach(async (to, from) => {
      if (to.matched.length) {
        if (to.matched[0].path !== "/") {
          redirectTo.value = null;
        }
        return;
      }
      if (from === START_LOCATION_NORMALIZED) {
        await ctx.$loader.initTask;
        to = this.router.resolve(to);
        if (to.matched.length) return to;
      }
      redirectTo.value = to.fullPath;
      const result = this.cache["home"] || "/";
      if (result === to.fullPath) return;
      return result;
    });
    this.slot({
      type: "global",
      component: Overlay
    });
  }
  slot(options) {
    options.order ?? (options.order = 0);
    options.component = this.ctx.wrapComponent(options.component);
    if (options.when) options.disabled = () => !options.when();
    return this.ctx.effect(() => {
      var _a15, _b;
      const list = (_a15 = this.views)[_b = options.type] || (_a15[_b] = []);
      insert(list, options);
      return () => {
        remove(list, options);
        if (!list.length) delete this.views[options.type];
      };
    });
  }
  page(options) {
    options.component = this.ctx.wrapComponent(options.component);
    return this.ctx.effect(() => {
      const activity = new Activity(this.ctx, options);
      return () => activity.dispose();
    });
  }
};

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/plugins/theme.ts
var preferDark = usePreferredDark();
var config2 = useConfig();
var colorMode = computed(() => {
  const mode = config2.value.theme.mode;
  if (mode !== "auto") return mode;
  return preferDark.value ? "dark" : "light";
});
var useColorMode = () => colorMode;
var ThemeService = class extends Service3 {
  constructor(ctx) {
    super(ctx, "$theme", true);
    ctx.mixin("$theme", ["theme"]);
    ctx.internal.themes = reactive({});
    ctx.settings({
      id: "appearance",
      title: "外观设置",
      order: 900,
      schema: lib_default.object({
        theme: lib_default.object({
          mode: lib_default.union([
            lib_default.const("auto").description("跟随系统"),
            lib_default.const("dark").description("深色"),
            lib_default.const("light").description("浅色")
          ]).default("auto").description("主题偏好。"),
          dark: lib_default.string().role("theme", { mode: "dark" }).default("default-dark").description("深色主题。"),
          light: lib_default.string().role("theme", { mode: "light" }).default("default-light").description("浅色主题。")
        }).description("主题设置")
      })
    });
    ctx.effect(() => watchEffect(() => {
      if (!config2.value.theme) return;
      const root2 = window.document.querySelector("html");
      root2.setAttribute("theme", config2.value.theme[colorMode.value]);
      if (colorMode.value === "dark") {
        root2.classList.add("dark");
      } else {
        root2.classList.remove("dark");
      }
    }, { flush: "post" }));
  }
  theme(options) {
    markRaw(options);
    for (const [type, component] of Object.entries(options.components || {})) {
      this.ctx.slot({
        type,
        disabled: () => config2.value.theme[colorMode.value] !== options.id,
        component
      });
    }
    return this.ctx.effect(() => {
      this.ctx.internal.themes[options.id] = options;
      return () => delete this.ctx.internal.themes[options.id];
    });
  }
};

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/context.ts
function useContext() {
  const parent = inject("cordis");
  const fork = parent.plugin(() => {
  });
  onBeforeUnmount(() => fork.dispose());
  return fork.ctx;
}
function useRpc() {
  var _a15;
  const parent = inject("cordis");
  return (_a15 = parent.extension) == null ? void 0 : _a15.data;
}
var Context4 = class extends Context2 {
  constructor() {
    super();
    __publicField(this, "app");
    this.extension = null;
    this.internal = {};
    this.app = createApp(defineComponent({
      setup: () => () => [
        h(resolveComponent("k-slot"), { name: "root", single: true }),
        h(resolveComponent("k-slot"), { name: "global" })
      ]
    }));
    this.app.provide("cordis", this);
    this.plugin(ActionService);
    this.plugin(I18nService);
    this.plugin(LoaderService);
    this.plugin(RouterService);
    this.plugin(SettingService);
    this.plugin(ThemeService);
    this.on("ready", async () => {
      await this.$loader.initTask;
      this.app.use(this.$i18n.i18n);
      this.app.use(this.$router.router);
      this.app.mount("#app");
    });
  }
  addEventListener(type, listener, options) {
    return this.effect(() => {
      window.addEventListener(type, listener, options);
      return () => window.removeEventListener(type, listener, options);
    });
  }
  wrapComponent(component) {
    if (!component) return;
    if (!this.extension) return component;
    return defineComponent((props, { slots }) => {
      provide("cordis", this);
      return () => h(component, props, slots);
    });
  }
};
markRaw(Context2.prototype);
markRaw(EffectScope.prototype);

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/link.ts
var KActivityLink = defineComponent({
  props: {
    id: String
  },
  setup(props, { slots }) {
    const ctx = useContext();
    return () => {
      const activity = ctx.$router.pages[props.id];
      return h(RouterLink, {
        to: ctx.$router.cache[activity == null ? void 0 : activity.id] || (activity == null ? void 0 : activity.path.replace(/:.+/, ""))
      }, {
        default: () => {
          var _a15;
          return ((_a15 = slots.default) == null ? void 0 : _a15.call(slots)) ?? (activity == null ? void 0 : activity.name);
        }
      });
    };
  }
});
function link_default(app) {
  app.component("k-activity-link", KActivityLink);
}

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/slot.ts
var KSlot = defineComponent({
  props: {
    name: String,
    data: Object,
    single: Boolean
  },
  setup(props, { slots }) {
    const ctx = useContext();
    return () => {
      var _a15, _b, _c;
      const internal = props.single ? [] : [...((_a15 = slots.default) == null ? void 0 : _a15.call(slots)) || []].filter((node) => node.type === KSlotItem).map((node) => {
        var _a16;
        return { node, order: ((_a16 = node.props) == null ? void 0 : _a16.order) || 0 };
      });
      const external = [...ctx.$router.views[props.name] || []].filter((item) => {
        var _a16;
        return !((_a16 = item.disabled) == null ? void 0 : _a16.call(item));
      }).map((item) => ({
        node: h(item.component, { ...props.data }, slots),
        order: item.order,
        layer: 1
      }));
      const children = [...internal, ...external].sort((a, b) => b.order - a.order);
      if (props.single) {
        return ((_b = children[0]) == null ? void 0 : _b.node) || ((_c = slots.default) == null ? void 0 : _c.call(slots));
      }
      return children.map((item) => item.node);
    };
  }
});
var KSlotItem = defineComponent({
  props: {
    order: Number
  },
  setup(props, { slots }) {
    return () => {
      var _a15;
      return (_a15 = slots.default) == null ? void 0 : _a15.call(slots);
    };
  }
});
function defineSlotComponent(name) {
  return defineComponent({
    inheritAttrs: false,
    setup(_, { slots, attrs }) {
      return () => h(KSlot, { name, data: attrs, single: true }, slots);
    }
  });
}
var slot_default = (app) => {
  app.component("k-slot", KSlot);
  app.component("k-slot-item", KSlotItem);
  app.component("k-layout", defineSlotComponent("layout"));
  app.component("k-status", defineSlotComponent("status"));
};

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/components/index.ts
import "D:/Python_Project/Moebot-NEXT/node_modules/.bun/element-plus@2.7.7+cc1f68d090f48582/node_modules/element-plus/dist/index.css";
var loading = ElLoading.service;
var message = ElMessage;
var messageBox = ElMessageBox;
src_default.extensions.add({
  type: "any",
  role: "dynamic",
  component: Dynamic
});
src_default.extensions.add({
  type: "array",
  role: "perms",
  component: Perms,
  validate: () => !!store.permissions
});
function components_default(app) {
  app.use(installer);
  app.component("k-markdown", src_default2);
  app.use(common_default);
  app.use(client_default);
  app.use(icons_exports);
  app.use(layout_default);
  app.use(link_default);
  app.use(slot_default);
}

// node_modules/.bun/@satorijs+protocol@1.7.0/node_modules/@satorijs/protocol/lib/index.mjs
var lib_exports3 = {};
__export(lib_exports3, {
  Channel: () => Channel,
  Methods: () => Methods,
  Opcode: () => Opcode,
  Resource: () => Resource,
  Status: () => Status,
  WebSocket: () => WebSocket,
  transformKey: () => transformKey
});

// node_modules/.bun/@satorijs+element@3.2.0/node_modules/@satorijs/element/lib/index.mjs
var __defProp8 = Object.defineProperty;
var __getOwnPropNames2 = Object.getOwnPropertyNames;
var __name8 = (target, value) => __defProp8(target, "name", { value, configurable: true });
var __commonJS3 = (cb, mod) => function __require() {
  return mod || (0, cb[__getOwnPropNames2(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
};
var require_index = __commonJS3({
  "src/index.ts"(exports, module) {
    var _a15;
    var kElement = Symbol.for("satori.element");
    var ElementConstructor = (_a15 = class {
      get data() {
        return this.attrs;
      }
      getTagName() {
        var _a16;
        if (this.type === "component") {
          return ((_a16 = this.attrs.is) == null ? void 0 : _a16.name) ?? "component";
        } else {
          return this.type;
        }
      }
      toAttrString() {
        return Object.entries(this.attrs).map(([key, value]) => {
          if (isNullable(value)) return "";
          key = hyphenate(key);
          if (value === true) return ` ${key}`;
          if (value === false) return ` no-${key}`;
          return ` ${key}="${Element.escape("" + value, true)}"`;
        }).join("");
      }
      toString(strip = false) {
        if (this.type === "text" && "content" in this.attrs) {
          return strip ? this.attrs.content : Element.escape(this.attrs.content);
        }
        const inner = this.children.map((child) => child.toString(strip)).join("");
        if (strip) return inner;
        const attrs = this.toAttrString();
        const tag = this.getTagName();
        if (!this.children.length) return `<${tag}${attrs}/>`;
        return `<${tag}${attrs}>${inner}</${tag}>`;
      }
    }, __name8(_a15, "ElementConstructor"), _a15);
    defineProperty(ElementConstructor, "name", "Element");
    defineProperty(ElementConstructor.prototype, kElement, true);
    function Element(type, ...args) {
      const el = Object.create(ElementConstructor.prototype);
      const attrs = {}, children = [];
      if (args[0] && typeof args[0] === "object" && !Element.isElement(args[0]) && !Array.isArray(args[0])) {
        const props = args.shift();
        for (const [key, value] of Object.entries(props)) {
          if (isNullable(value)) continue;
          if (key === "children") {
            args.push(...makeArray(value));
          } else {
            attrs[camelize(key)] = value;
          }
        }
      }
      for (const child of args) {
        children.push(...Element.toElementArray(child));
      }
      if (typeof type === "function") {
        attrs.is = type;
        type = "component";
      }
      return Object.assign(el, { type, attrs, children });
    }
    __name8(Element, "Element");
    var evaluate = new Function("expr", "context", `
  try {
    with (context) {
      return eval(expr)
    }
  } catch {}
`);
    ((Element2) => {
      Element2.jsx = Element2;
      Element2.jsxs = Element2;
      Element2.jsxDEV = Element2;
      Element2.Fragment = "template";
      function isElement(source) {
        return source && typeof source === "object" && source[kElement];
      }
      Element2.isElement = isElement;
      __name8(isElement, "isElement");
      function toElement(content) {
        if (typeof content === "string" || typeof content === "number" || typeof content === "boolean") {
          content = "" + content;
          if (content) return Element2("text", { content });
        } else if (isElement(content)) {
          return content;
        } else if (!isNullable(content)) {
          throw new TypeError(`Invalid content: ${content}`);
        }
      }
      Element2.toElement = toElement;
      __name8(toElement, "toElement");
      function toElementArray(content) {
        if (Array.isArray(content)) {
          return content.map(toElement).filter(isNonNullable);
        } else {
          return [toElement(content)].filter(isNonNullable);
        }
      }
      Element2.toElementArray = toElementArray;
      __name8(toElementArray, "toElementArray");
      function normalize(source, context) {
        return typeof source === "string" ? parse(source, context) : toElementArray(source);
      }
      Element2.normalize = normalize;
      __name8(normalize, "normalize");
      function escape(source, inline = false) {
        const result = (source ?? "").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
        return inline ? result.replace(/"/g, "&quot;") : result;
      }
      Element2.escape = escape;
      __name8(escape, "escape");
      function unescape(source) {
        return source.replace(/&lt;/g, "<").replace(/&gt;/g, ">").replace(/&quot;/g, '"').replace(/&#(\d+);/g, (_, code) => code === "38" ? _ : String.fromCharCode(+code)).replace(/&#x([0-9a-f]+);/gi, (_, code) => code === "26" ? _ : String.fromCharCode(parseInt(code, 16))).replace(/&(amp|#38|#x26);/g, "&");
      }
      Element2.unescape = unescape;
      __name8(unescape, "unescape");
      function from(source, options = {}) {
        var _a16;
        const elements = parse(source);
        if (options.caret) {
          if (options.type && ((_a16 = elements[0]) == null ? void 0 : _a16.type) !== options.type) return;
          return elements[0];
        }
        return select(elements, options.type || "*")[0];
      }
      Element2.from = from;
      __name8(from, "from");
      const combRegExp = / *([ >+~]) */g;
      function parseSelector(input) {
        return input.split(",").map((query) => {
          const selectors = [];
          query = query.trim();
          let combCap, combinator = " ";
          while (combCap = combRegExp.exec(query)) {
            selectors.push({ type: query.slice(0, combCap.index), combinator });
            combinator = combCap[1];
            query = query.slice(combCap.index + combCap[0].length);
          }
          selectors.push({ type: query, combinator });
          return selectors;
        });
      }
      Element2.parseSelector = parseSelector;
      __name8(parseSelector, "parseSelector");
      function select(source, query) {
        if (!source || !query) return [];
        if (typeof source === "string") source = parse(source);
        if (typeof query === "string") query = parseSelector(query);
        if (!query.length) return [];
        let adjacent = [];
        const results = [];
        for (const [index, element] of source.entries()) {
          const inner = [];
          const local = [...query, ...adjacent];
          adjacent = [];
          let matched = false;
          for (const group of local) {
            const { type, combinator } = group[0];
            if (type === element.type || type === "*") {
              if (group.length === 1) {
                matched = true;
              } else if ([" ", ">"].includes(group[1].combinator)) {
                inner.push(group.slice(1));
              } else if (group[1].combinator === "+") {
                adjacent.push(group.slice(1));
              } else {
                query.push(group.slice(1));
              }
            }
            if (combinator === " ") {
              inner.push(group);
            }
          }
          if (matched) results.push(source[index]);
          results.push(...select(element.children, inner));
        }
        return results;
      }
      Element2.select = select;
      __name8(select, "select");
      function interpolate(expr, context) {
        expr = expr.trim();
        if (!/^[\w.]+$/.test(expr)) {
          return evaluate(expr, context) ?? "";
        }
        let value = context;
        for (const part of expr.split(".")) {
          value = value[part];
          if (isNullable(value)) return "";
        }
        return value ?? "";
      }
      Element2.interpolate = interpolate;
      __name8(interpolate, "interpolate");
      const tagRegExp1 = /(?<comment><!--[\s\S]*?-->)|(?<tag><(\/?)([^!\s>/]*)([^>]*?)\s*(\/?)>)/;
      const tagRegExp2 = /(?<comment><!--[\s\S]*?-->)|(?<tag><(\/?)([^!\s>/]*)([^>]*?)\s*(\/?)>)|(?<curly>\{(?<derivative>[@:/#][^\s}]*)?[\s\S]*?\})/;
      const attrRegExp1 = /([^\s=]+)(?:="(?<value1>[^"]*)"|='(?<value2>[^']*)')?/g;
      const attrRegExp2 = /([^\s=]+)(?:="(?<value1>[^"]*)"|='(?<value2>[^']*)'|=\{(?<curly>[^}]+)\})?/g;
      let Position;
      ((Position2) => {
        Position2[Position2["OPEN"] = 0] = "OPEN";
        Position2[Position2["CLOSE"] = 1] = "CLOSE";
        Position2[Position2["EMPTY"] = 2] = "EMPTY";
        Position2[Position2["CONTINUE"] = 3] = "CONTINUE";
      })(Position || (Position = {}));
      function parse(source, context) {
        const tokens = [];
        function pushText(content) {
          if (content) tokens.push(content);
        }
        __name8(pushText, "pushText");
        const tagRegExp = context ? tagRegExp2 : tagRegExp1;
        let tagCap;
        let trimStart = true;
        while (tagCap = tagRegExp.exec(source)) {
          const { curly, comment, derivative } = tagCap.groups;
          const trimEnd = !curly;
          parseContent(source.slice(0, tagCap.index), trimStart, trimEnd);
          trimStart = trimEnd;
          source = source.slice(tagCap.index + tagCap[0].length);
          const [_, , , close, type, extra, empty] = tagCap;
          if (comment) continue;
          if (curly) {
            let name = "", position = 2;
            if (derivative) {
              name = derivative.slice(1);
              position = {
                "@": 2,
                "#": 0,
                "/": 1,
                ":": 3
                /* CONTINUE */
              }[derivative[0]];
            }
            tokens.push({
              type: "curly",
              name,
              position,
              source: curly,
              extra: curly.slice(1 + (derivative ?? "").length, -1)
            });
            continue;
          }
          tokens.push({
            type: "angle",
            source: _,
            name: type || Element2.Fragment,
            position: close ? 1 : empty ? 2 : 0,
            extra
          });
        }
        parseContent(source, trimStart, true);
        function parseContent(source2, trimStart2, trimEnd) {
          source2 = unescape(source2);
          if (trimStart2) source2 = source2.replace(/^\s*\n\s*/, "");
          if (trimEnd) source2 = source2.replace(/\s*\n\s*$/, "");
          pushText(source2);
        }
        __name8(parseContent, "parseContent");
        return parseTokens(foldTokens(tokens), context);
      }
      Element2.parse = parse;
      __name8(parse, "parse");
      function foldTokens(tokens) {
        const stack = [[{
          type: "angle",
          name: Element2.Fragment,
          position: 0,
          source: "",
          extra: "",
          children: { default: [] }
        }, "default"]];
        function pushToken(...tokens2) {
          const [token, slot] = stack[0];
          token.children[slot].push(...tokens2);
        }
        __name8(pushToken, "pushToken");
        for (const token of tokens) {
          if (typeof token === "string") {
            pushToken(token);
            continue;
          }
          const { name, position } = token;
          if (position === 1) {
            if (stack[0][0].name === name) {
              stack.shift();
            }
          } else if (position === 3) {
            stack[0][0].children[name] = [];
            stack[0][1] = name;
          } else if (position === 0) {
            pushToken(token);
            token.children = { default: [] };
            stack.unshift([token, "default"]);
          } else {
            pushToken(token);
          }
        }
        return stack[stack.length - 1][0].children.default;
      }
      __name8(foldTokens, "foldTokens");
      function parseTokens(tokens, context) {
        const result = [];
        for (const token of tokens) {
          if (typeof token === "string") {
            result.push(Element2("text", { content: token }));
          } else if (token.type === "angle") {
            const attrs = {};
            const attrRegExp = context ? attrRegExp2 : attrRegExp1;
            let attrCap;
            while (attrCap = attrRegExp.exec(token.extra)) {
              const [, key, v1, v2 = v1, v3] = attrCap;
              if (v3) {
                attrs[key] = interpolate(v3, context);
              } else if (!isNullable(v2)) {
                attrs[key] = unescape(v2);
              } else if (key.startsWith("no-")) {
                attrs[key.slice(3)] = false;
              } else {
                attrs[key] = true;
              }
            }
            result.push(Element2(token.name, attrs, token.children && parseTokens(token.children.default, context)));
          } else if (!token.name) {
            result.push(...toElementArray(interpolate(token.extra, context)));
          } else if (token.name === "if") {
            if (evaluate(token.extra, context)) {
              result.push(...parseTokens(token.children.default, context));
            } else {
              result.push(...parseTokens(token.children.else || [], context));
            }
          } else if (token.name === "each") {
            const [expr, ident] = token.extra.split(/\s+as\s+/);
            const items = interpolate(expr, context);
            if (!items || !items[Symbol.iterator]) continue;
            for (const item of items) {
              result.push(...parseTokens(token.children.default, { ...context, [ident]: item }));
            }
          }
        }
        return result;
      }
      __name8(parseTokens, "parseTokens");
      function visit(element, rules, session) {
        const { type, attrs, children } = element;
        if (typeof rules === "function") {
          return rules(element, session);
        } else {
          let result = rules[typeof type === "string" ? type : ""] ?? rules.default ?? true;
          if (typeof result === "function") {
            result = result(attrs, children, session);
          }
          return result;
        }
      }
      __name8(visit, "visit");
      function transform(source, rules, session) {
        const elements = typeof source === "string" ? parse(source) : source;
        const output = [];
        elements.forEach((element) => {
          const { type, attrs, children } = element;
          const result = visit(element, rules, session);
          if (result === true) {
            output.push(Element2(type, attrs, transform(children, rules, session)));
          } else if (result !== false) {
            output.push(...toElementArray(result));
          }
        });
        return typeof source === "string" ? output.join("") : output;
      }
      Element2.transform = transform;
      __name8(transform, "transform");
      async function transformAsync(source, rules, session) {
        const elements = typeof source === "string" ? parse(source) : source;
        const children = (await Promise.all(elements.map(async (element) => {
          const { type, attrs, children: children2 } = element;
          const result = await visit(element, rules, session);
          if (result === true) {
            return [Element2(type, attrs, await transformAsync(children2, rules, session))];
          } else if (result !== false) {
            return toElementArray(result);
          } else {
            return [];
          }
        }))).flat(1);
        return typeof source === "string" ? children.join("") : children;
      }
      Element2.transformAsync = transformAsync;
      __name8(transformAsync, "transformAsync");
      function createFactory(type, ...keys) {
        return (...args) => {
          const element = Element2(type);
          keys.forEach((key, index) => {
            if (!isNullable(args[index])) {
              element.attrs[key] = args[index];
            }
          });
          if (args[keys.length]) {
            Object.assign(element.attrs, args[keys.length]);
          }
          return element;
        };
      }
      __name8(createFactory, "createFactory");
      Element2.warn = __name8(() => {
      }, "warn");
      function createAssetFactory(type) {
        return (src, ...args) => {
          let prefix = "base64://";
          if (typeof args[0] === "string") {
            prefix = `data:${args.shift()};base64,`;
          }
          if (is("Buffer", src)) {
            src = prefix + src.toString("base64");
          } else if (is("ArrayBuffer", src)) {
            src = prefix + Binary.toBase64(src);
          } else if (ArrayBuffer.isView(src)) {
            src = prefix + Binary.toBase64(src.buffer);
          }
          if (src.startsWith("base64://")) {
            (0, Element2.warn)(`protocol "base64:" is deprecated and will be removed in the future, please use "data:" instead`);
          }
          return Element2(type, { ...args[0], src });
        };
      }
      __name8(createAssetFactory, "createAssetFactory");
      Element2.text = createFactory("text", "content");
      Element2.at = createFactory("at", "id");
      Element2.sharp = createFactory("sharp", "id");
      Element2.quote = createFactory("quote", "id");
      Element2.emoji = createFactory("emoji", "id");
      Element2.image = createAssetFactory("img");
      Element2.img = createAssetFactory("img");
      Element2.video = createAssetFactory("video");
      Element2.audio = createAssetFactory("audio");
      Element2.file = createAssetFactory("file");
      function i18n(path, children) {
        return Element2("i18n", typeof path === "string" ? { path } : path, children);
      }
      Element2.i18n = i18n;
      __name8(i18n, "i18n");
    })(Element || (Element = {}));
    module.exports = Element;
  }
});
var lib_default2 = require_index();

// node_modules/.bun/@satorijs+protocol@1.7.0/node_modules/@satorijs/protocol/lib/index.mjs
var __defProp9 = Object.defineProperty;
var __name9 = (target, value) => __defProp9(target, "name", { value, configurable: true });
function Field(name) {
  return { name };
}
__name9(Field, "Field");
function Method(name, fields, isForm = false) {
  return { name, fields: fields.map(Field), isForm };
}
__name9(Method, "Method");
var Methods = {
  "channel.get": Method("getChannel", ["channel_id", "guild_id"]),
  "channel.list": Method("getChannelList", ["guild_id", "next"]),
  "channel.create": Method("createChannel", ["guild_id", "data"]),
  "channel.update": Method("updateChannel", ["channel_id", "data"]),
  "channel.delete": Method("deleteChannel", ["channel_id"]),
  "channel.mute": Method("muteChannel", ["channel_id", "guild_id", "enable"]),
  "message.create": Method("createMessage", ["channel_id", "content", "referrer"]),
  "message.update": Method("editMessage", ["channel_id", "message_id", "content"]),
  "message.delete": Method("deleteMessage", ["channel_id", "message_id"]),
  "message.get": Method("getMessage", ["channel_id", "message_id"]),
  "message.list": Method("getMessageList", ["channel_id", "next", "direction", "limit", "order"]),
  "reaction.create": Method("createReaction", ["channel_id", "message_id", "emoji_id"]),
  "reaction.delete": Method("deleteReaction", ["channel_id", "message_id", "emoji_id", "user_id"]),
  "reaction.clear": Method("clearReaction", ["channel_id", "message_id", "emoji_id"]),
  "reaction.list": Method("getReactionList", ["channel_id", "message_id", "emoji_id", "next"]),
  "upload.create": Method("createUpload", [], true),
  "guild.get": Method("getGuild", ["guild_id"]),
  "guild.list": Method("getGuildList", ["next"]),
  "guild.member.get": Method("getGuildMember", ["guild_id", "user_id"]),
  "guild.member.list": Method("getGuildMemberList", ["guild_id", "next"]),
  "guild.member.kick": Method("kickGuildMember", ["guild_id", "user_id", "permanent"]),
  "guild.member.mute": Method("muteGuildMember", ["guild_id", "user_id", "duration", "reason"]),
  "guild.member.role.set": Method("setGuildMemberRole", ["guild_id", "user_id", "role_id"]),
  "guild.member.role.unset": Method("unsetGuildMemberRole", ["guild_id", "user_id", "role_id"]),
  "guild.member.role.list": Method("getGuildMemberRoleList", ["guild_id", "user_id", "next"]),
  "guild.role.list": Method("getGuildRoleList", ["guild_id", "next"]),
  "guild.role.create": Method("createGuildRole", ["guild_id", "data"]),
  "guild.role.update": Method("updateGuildRole", ["guild_id", "role_id", "data"]),
  "guild.role.delete": Method("deleteGuildRole", ["guild_id", "role_id"]),
  "login.get": Method("getLogin", []),
  "user.get": Method("getUser", ["user_id"]),
  "user.channel.create": Method("createDirectChannel", ["user_id", "guild_id"]),
  "friend.list": Method("getFriendList", ["next"]),
  "friend.delete": Method("deleteFriend", ["user_id"]),
  "friend.approve": Method("handleFriendRequest", ["message_id", "approve", "comment"]),
  "guild.approve": Method("handleGuildRequest", ["message_id", "approve", "comment"]),
  "guild.member.approve": Method("handleGuildMemberRequest", ["message_id", "approve", "comment"])
};
var Channel;
((Channel2) => {
  let Type;
  ((Type2) => {
    Type2[Type2["TEXT"] = 0] = "TEXT";
    Type2[Type2["DIRECT"] = 1] = "DIRECT";
    Type2[Type2["CATEGORY"] = 2] = "CATEGORY";
    Type2[Type2["VOICE"] = 3] = "VOICE";
  })(Type = Channel2.Type || (Channel2.Type = {}));
})(Channel || (Channel = {}));
function Resource(attrs = [], children = [], content) {
  return { attrs, children, content };
}
__name9(Resource, "Resource");
((Resource2) => {
  const Definitions = {
    user: Resource2(["id", "name", "nick", "avatar", "isBot"]),
    member: Resource2(["name", "nick", "avatar"]),
    channel: Resource2(["id", "type", "name"]),
    guild: Resource2(["id", "name", "avatar"]),
    quote: Resource2(["id"], ["quote", "user", "member", "channel"], "content")
  };
  function encode(type, data) {
    const resource = Definitions[type];
    const element = lib_default2(type, pick(data, resource.attrs));
    for (const key of resource.children) {
      if (isNullable(data[key])) continue;
      element.children.push(encode(key, data[key]));
    }
    if (resource.content && !isNullable(data[resource.content])) {
      element.children.push(...lib_default2.parse(data[resource.content]));
    }
    return element;
  }
  Resource2.encode = encode;
  __name9(encode, "encode");
  function decode(element) {
    const data = element.attrs;
    const resource = Definitions[element.type];
    for (const key of resource.children) {
      const index = element.children.findIndex((el) => el.type === key);
      if (index === -1) continue;
      const [child] = element.children.splice(index, 1);
      data[key] = decode(child);
    }
    if (resource.content && element.children.length) {
      data[resource.content] = element.children.join("");
    }
    return data;
  }
  Resource2.decode = decode;
  __name9(decode, "decode");
})(Resource || (Resource = {}));
function transformKey(source, callback) {
  if (!source || typeof source !== "object") return source;
  if (Array.isArray(source)) return source.map((value) => transformKey(value, callback));
  return Object.fromEntries(Object.entries(source).map(([key, value]) => {
    if (key.startsWith("_") || key === "referrer") return [key, value];
    return [callback(key), transformKey(value, callback)];
  }));
}
__name9(transformKey, "transformKey");
var Status = ((Status2) => {
  Status2[Status2["OFFLINE"] = 0] = "OFFLINE";
  Status2[Status2["ONLINE"] = 1] = "ONLINE";
  Status2[Status2["CONNECT"] = 2] = "CONNECT";
  Status2[Status2["DISCONNECT"] = 3] = "DISCONNECT";
  Status2[Status2["RECONNECT"] = 4] = "RECONNECT";
  return Status2;
})(Status || {});
var Opcode = ((Opcode2) => {
  Opcode2[Opcode2["EVENT"] = 0] = "EVENT";
  Opcode2[Opcode2["PING"] = 1] = "PING";
  Opcode2[Opcode2["PONG"] = 2] = "PONG";
  Opcode2[Opcode2["IDENTIFY"] = 3] = "IDENTIFY";
  Opcode2[Opcode2["READY"] = 4] = "READY";
  Opcode2[Opcode2["META"] = 5] = "META";
  return Opcode2;
})(Opcode || {});
var WebSocket;
((WebSocket2) => {
  WebSocket2.CONNECTING = 0;
  WebSocket2.OPEN = 1;
  WebSocket2.CLOSING = 2;
  WebSocket2.CLOSED = 3;
})(WebSocket || (WebSocket = {}));

// node_modules/.bun/@koishijs+client@5.30.11+2dc76b59b67e6600/node_modules/@koishijs/client/client/index.ts
var client_default2 = components_default;
var root = new Context4();
root.app.use(components_default);
root.on("activity", (data) => !data);
var router = root.$router.router;
var activities = root.$router.pages;
export {
  Activity,
  Binary,
  ChatImage,
  Context4 as Context,
  DatetimeFormat,
  I18nD,
  I18nInjectionKey,
  I18nN,
  I18nT,
  IconAdd,
  IconArrowDown,
  IconArrowUp,
  IconBranch,
  IconClose,
  IconCode,
  IconCollapse,
  IconDelete,
  IconEllipsis,
  IconExpand,
  IconExternal,
  IconEye,
  IconEyeSlash,
  IconInsertAfter,
  IconInsertBefore,
  IconInvalid,
  IconRedo,
  IconReset,
  IconSquareCheck,
  IconSquareEmpty,
  IconUndo,
  KSlot,
  NumberFormat,
  Primitive,
  lib_exports3 as Satori,
  lib_default as Schema,
  src_default as SchemaBase,
  ScopeStatus,
  Service3 as Service,
  Time,
  Translation,
  lib_exports3 as Universal,
  VERSION,
  VirtualList,
  activities,
  arrayBufferToBase64,
  arrayBufferToHex,
  base64ToArrayBuffer,
  camelCase,
  camelize,
  capitalize,
  castToVueI18n,
  clone,
  connect,
  contain,
  createI18n,
  createStorage,
  deduplicate,
  deepEqual,
  client_default2 as default,
  defineExtension,
  defineProperty,
  difference,
  filterKeys,
  form,
  formatProperty,
  global2 as global,
  hexToArrayBuffer,
  hyphenate,
  icons_exports as icons,
  intersection,
  is,
  isNonNullable,
  isNullable,
  isPlainObject,
  loading,
  makeArray,
  mapValues,
  message,
  messageBox,
  noop,
  omit,
  original,
  paramCase,
  pick,
  provideStorage,
  receive,
  redirectTo,
  remove,
  resolved,
  root,
  router,
  sanitize,
  send,
  snakeCase,
  socket,
  store,
  trimSlash,
  uncapitalize,
  union,
  unwrapExports,
  useColorMode,
  useConfig,
  useContext,
  useI18n,
  useI18nText,
  useMenu,
  useRpc,
  useStorage,
  vTDirective,
  mapValues as valueMap,
  withProxy
};
//# sourceMappingURL=@koishijs_client.js.map
