import {
  fallbackWithLocaleChain,
  useI18n
} from "./chunk-OYU45GEP.js";
import {
  computed,
  getCurrentInstance,
  ref,
  watch
} from "./chunk-QBIFJQ73.js";
import {
  __publicField
} from "./chunk-Q4XP6UTR.js";

// node_modules/.bun/cosmokit@1.8.1/node_modules/cosmokit/lib/index.mjs
function noop() {
}
function isNullable(value) {
  return value === null || value === void 0;
}
function isNonNullable(value) {
  return !isNullable(value);
}
function isPlainObject(data) {
  return data && typeof data === "object" && !Array.isArray(data);
}
function filterKeys(object, filter) {
  return Object.fromEntries(Object.entries(object).filter(([key, value]) => filter(key, value)));
}
function mapValues(object, transform) {
  return Object.fromEntries(Object.entries(object).map(([key, value]) => [key, transform(value, key)]));
}
function pick(source, keys, forced) {
  if (!keys) return { ...source };
  const result = {};
  for (const key of keys) {
    if (forced || source[key] !== void 0) result[key] = source[key];
  }
  return result;
}
function omit(source, keys) {
  if (!keys) return { ...source };
  const result = { ...source };
  for (const key of keys) {
    Reflect.deleteProperty(result, key);
  }
  return result;
}
function defineProperty(object, key, value) {
  return Object.defineProperty(object, key, { writable: true, value, enumerable: false });
}
function contain(array1, array2) {
  return array2.every((item) => array1.includes(item));
}
function intersection(array1, array2) {
  return array1.filter((item) => array2.includes(item));
}
function difference(array1, array2) {
  return array1.filter((item) => !array2.includes(item));
}
function union(array1, array2) {
  return Array.from(/* @__PURE__ */ new Set([...array1, ...array2]));
}
function deduplicate(array) {
  return [...new Set(array)];
}
function remove(list, item) {
  const index = list == null ? void 0 : list.indexOf(item);
  if (index >= 0) {
    list.splice(index, 1);
    return true;
  } else {
    return false;
  }
}
function makeArray(source) {
  return Array.isArray(source) ? source : isNullable(source) ? [] : [source];
}
function is(type, value) {
  if (arguments.length === 1) return (value2) => is(type, value2);
  return type in globalThis && value instanceof globalThis[type] || Object.prototype.toString.call(value).slice(8, -1) === type;
}
function isArrayBufferLike(value) {
  return is("ArrayBuffer", value) || is("SharedArrayBuffer", value);
}
function isArrayBufferSource(value) {
  return isArrayBufferLike(value) || ArrayBuffer.isView(value);
}
var Binary;
((Binary2) => {
  Binary2.is = isArrayBufferLike;
  Binary2.isSource = isArrayBufferSource;
  function fromSource(source) {
    if (ArrayBuffer.isView(source)) {
      return source.buffer.slice(source.byteOffset, source.byteOffset + source.byteLength);
    } else {
      return source;
    }
  }
  Binary2.fromSource = fromSource;
  function toBase64(source) {
    source = fromSource(source);
    if (typeof Buffer !== "undefined") {
      return Buffer.from(source).toString("base64");
    }
    let binary = "";
    const bytes = new Uint8Array(source);
    for (let i = 0; i < bytes.byteLength; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  }
  Binary2.toBase64 = toBase64;
  function fromBase64(source) {
    if (typeof Buffer !== "undefined") return fromSource(Buffer.from(source, "base64"));
    return Uint8Array.from(atob(source), (c) => c.charCodeAt(0));
  }
  Binary2.fromBase64 = fromBase64;
  function toHex(source) {
    source = fromSource(source);
    if (typeof Buffer !== "undefined") return Buffer.from(source).toString("hex");
    return Array.from(new Uint8Array(source), (byte) => byte.toString(16).padStart(2, "0")).join("");
  }
  Binary2.toHex = toHex;
  function fromHex(source) {
    if (typeof Buffer !== "undefined") return fromSource(Buffer.from(source, "hex"));
    const hex = source.length % 2 === 0 ? source : source.slice(0, source.length - 1);
    const buffer = [];
    for (let i = 0; i < hex.length; i += 2) {
      buffer.push(parseInt(`${hex[i]}${hex[i + 1]}`, 16));
    }
    return Uint8Array.from(buffer).buffer;
  }
  Binary2.fromHex = fromHex;
})(Binary || (Binary = {}));
var base64ToArrayBuffer = Binary.fromBase64;
var arrayBufferToBase64 = Binary.toBase64;
var hexToArrayBuffer = Binary.fromHex;
var arrayBufferToHex = Binary.toHex;
function clone(source, refs = /* @__PURE__ */ new Map()) {
  if (!source || typeof source !== "object") return source;
  if (is("Date", source)) return new Date(source.valueOf());
  if (is("RegExp", source)) return new RegExp(source.source, source.flags);
  if (isArrayBufferLike(source)) return source.slice(0);
  if (ArrayBuffer.isView(source)) return source.buffer.slice(source.byteOffset, source.byteOffset + source.byteLength);
  const cached = refs.get(source);
  if (cached) return cached;
  if (Array.isArray(source)) {
    const result2 = [];
    refs.set(source, result2);
    source.forEach((value, index) => {
      result2[index] = Reflect.apply(clone, null, [value, refs]);
    });
    return result2;
  }
  const result = Object.create(Object.getPrototypeOf(source));
  refs.set(source, result);
  for (const key of Reflect.ownKeys(source)) {
    const descriptor = { ...Reflect.getOwnPropertyDescriptor(source, key) };
    if ("value" in descriptor) {
      descriptor.value = Reflect.apply(clone, null, [descriptor.value, refs]);
    }
    Reflect.defineProperty(result, key, descriptor);
  }
  return result;
}
function deepEqual(a, b, strict) {
  if (a === b) return true;
  if (!strict && isNullable(a) && isNullable(b)) return true;
  if (typeof a !== typeof b) return false;
  if (typeof a !== "object") return false;
  if (!a || !b) return false;
  function check(test, then) {
    return test(a) ? test(b) ? then(a, b) : false : test(b) ? false : void 0;
  }
  return check(Array.isArray, (a2, b2) => a2.length === b2.length && a2.every((item, index) => deepEqual(item, b2[index]))) ?? check(is("Date"), (a2, b2) => a2.valueOf() === b2.valueOf()) ?? check(is("RegExp"), (a2, b2) => a2.source === b2.source && a2.flags === b2.flags) ?? check(isArrayBufferLike, (a2, b2) => {
    if (a2.byteLength !== b2.byteLength) return false;
    const viewA = new Uint8Array(a2);
    const viewB = new Uint8Array(b2);
    for (let i = 0; i < viewA.length; i++) {
      if (viewA[i] !== viewB[i]) return false;
    }
    return true;
  }) ?? Object.keys({ ...a, ...b }).every((key) => deepEqual(a[key], b[key], strict));
}
function capitalize(source) {
  return source.charAt(0).toUpperCase() + source.slice(1);
}
function uncapitalize(source) {
  return source.charAt(0).toLowerCase() + source.slice(1);
}
function camelCase(source) {
  return source.replace(/[_-][a-z]/g, (str) => str.slice(1).toUpperCase());
}
function tokenize(source, delimiters, delimiter) {
  const output = [];
  let state = 0;
  for (let i = 0; i < source.length; i++) {
    const code = source.charCodeAt(i);
    if (code >= 65 && code <= 90) {
      if (state === 1) {
        const next = source.charCodeAt(i + 1);
        if (next >= 97 && next <= 122) {
          output.push(delimiter);
        }
        output.push(code + 32);
      } else {
        if (state !== 0) {
          output.push(delimiter);
        }
        output.push(code + 32);
      }
      state = 1;
    } else if (code >= 97 && code <= 122) {
      output.push(code);
      state = 2;
    } else if (delimiters.includes(code)) {
      if (state !== 0) {
        output.push(delimiter);
      }
      state = 0;
    } else {
      output.push(code);
    }
  }
  return String.fromCharCode(...output);
}
function paramCase(source) {
  return tokenize(source, [45, 95], 45);
}
function snakeCase(source) {
  return tokenize(source, [45, 95], 95);
}
var camelize = camelCase;
var hyphenate = paramCase;
function formatProperty(key) {
  if (typeof key !== "string") return `[${key.toString()}]`;
  return /^[a-z_$][\w$]*$/i.test(key) ? `.${key}` : `[${JSON.stringify(key)}]`;
}
function trimSlash(source) {
  return source.replace(/\/$/, "");
}
function sanitize(source) {
  if (!source.startsWith("/")) source = "/" + source;
  return trimSlash(source);
}
var Time;
((Time2) => {
  Time2.millisecond = 1;
  Time2.second = 1e3;
  Time2.minute = Time2.second * 60;
  Time2.hour = Time2.minute * 60;
  Time2.day = Time2.hour * 24;
  Time2.week = Time2.day * 7;
  let timezoneOffset = (/* @__PURE__ */ new Date()).getTimezoneOffset();
  function setTimezoneOffset(offset) {
    timezoneOffset = offset;
  }
  Time2.setTimezoneOffset = setTimezoneOffset;
  function getTimezoneOffset() {
    return timezoneOffset;
  }
  Time2.getTimezoneOffset = getTimezoneOffset;
  function getDateNumber(date = /* @__PURE__ */ new Date(), offset) {
    if (typeof date === "number") date = new Date(date);
    if (offset === void 0) offset = timezoneOffset;
    return Math.floor((date.valueOf() / Time2.minute - offset) / 1440);
  }
  Time2.getDateNumber = getDateNumber;
  function fromDateNumber(value, offset) {
    const date = new Date(value * Time2.day);
    if (offset === void 0) offset = timezoneOffset;
    return new Date(+date + offset * Time2.minute);
  }
  Time2.fromDateNumber = fromDateNumber;
  const numeric = /\d+(?:\.\d+)?/.source;
  const timeRegExp = new RegExp(`^${[
    "w(?:eek(?:s)?)?",
    "d(?:ay(?:s)?)?",
    "h(?:our(?:s)?)?",
    "m(?:in(?:ute)?(?:s)?)?",
    "s(?:ec(?:ond)?(?:s)?)?"
  ].map((unit) => `(${numeric}${unit})?`).join("")}$`);
  function parseTime(source) {
    const capture = timeRegExp.exec(source);
    if (!capture) return 0;
    return (parseFloat(capture[1]) * Time2.week || 0) + (parseFloat(capture[2]) * Time2.day || 0) + (parseFloat(capture[3]) * Time2.hour || 0) + (parseFloat(capture[4]) * Time2.minute || 0) + (parseFloat(capture[5]) * Time2.second || 0);
  }
  Time2.parseTime = parseTime;
  function parseDate(date) {
    const parsed = parseTime(date);
    if (parsed) {
      date = Date.now() + parsed;
    } else if (/^\d{1,2}(:\d{1,2}){1,2}$/.test(date)) {
      date = `${(/* @__PURE__ */ new Date()).toLocaleDateString()}-${date}`;
    } else if (/^\d{1,2}-\d{1,2}-\d{1,2}(:\d{1,2}){1,2}$/.test(date)) {
      date = `${(/* @__PURE__ */ new Date()).getFullYear()}-${date}`;
    }
    return date ? new Date(date) : /* @__PURE__ */ new Date();
  }
  Time2.parseDate = parseDate;
  function format(ms) {
    const abs = Math.abs(ms);
    if (abs >= Time2.day - Time2.hour / 2) {
      return Math.round(ms / Time2.day) + "d";
    } else if (abs >= Time2.hour - Time2.minute / 2) {
      return Math.round(ms / Time2.hour) + "h";
    } else if (abs >= Time2.minute - Time2.second / 2) {
      return Math.round(ms / Time2.minute) + "m";
    } else if (abs >= Time2.second) {
      return Math.round(ms / Time2.second) + "s";
    }
    return ms + "ms";
  }
  Time2.format = format;
  function toDigits(source, length = 2) {
    return source.toString().padStart(length, "0");
  }
  Time2.toDigits = toDigits;
  function template(template2, time = /* @__PURE__ */ new Date()) {
    return template2.replace("yyyy", time.getFullYear().toString()).replace("yy", time.getFullYear().toString().slice(2)).replace("MM", toDigits(time.getMonth() + 1)).replace("dd", toDigits(time.getDate())).replace("hh", toDigits(time.getHours())).replace("mm", toDigits(time.getMinutes())).replace("ss", toDigits(time.getSeconds())).replace("SSS", toDigits(time.getMilliseconds(), 3));
  }
  Time2.template = template;
})(Time || (Time = {}));

// node_modules/.bun/schemastery@3.18.0/node_modules/schemastery/lib/index.mjs
var __defProp = Object.defineProperty;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __name = (target, value) => __defProp(target, "name", { value, configurable: true });
var __commonJS = (cb, mod) => function __require() {
  return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
};
var require_index = __commonJS({
  "src/index.ts"(exports, module) {
    var _a;
    var kSchema = Symbol.for("schemastery");
    var kValidationError = Symbol.for("ValidationError");
    globalThis.__schemastery_index__ ?? (globalThis.__schemastery_index__ = 0);
    globalThis.__schemastery_refs__ = void 0;
    var ValidationError = (_a = class extends TypeError {
      constructor(message, options) {
        let prefix = "$";
        for (const segment of options.path || []) {
          if (typeof segment === "string") {
            prefix += "." + segment;
          } else if (typeof segment === "number") {
            prefix += "[" + segment + "]";
          } else if (typeof segment === "symbol") {
            prefix += `[Symbol(${segment.toString()})]`;
          }
        }
        if (prefix.startsWith(".")) prefix = prefix.slice(1);
        super((prefix === "$" ? "" : `${prefix} `) + message);
        __publicField(this, "name", "ValidationError");
        this.options = options;
      }
      static is(error) {
        return !!(error == null ? void 0 : error[kValidationError]);
      }
    }, __name(_a, "ValidationError"), _a);
    Object.defineProperty(ValidationError.prototype, kValidationError, {
      value: true
    });
    var Schema2 = __name(function(options) {
      const schema = __name(function(data, options2 = {}) {
        return Schema2.resolve(data, schema, options2)[0];
      }, "schema");
      if (options.refs) {
        const refs = mapValues(options.refs, (options2) => new Schema2(options2));
        const getRef = __name((uid) => refs[uid], "getRef");
        for (const key in refs) {
          const options2 = refs[key];
          options2.sKey = getRef(options2.sKey);
          options2.inner = getRef(options2.inner);
          options2.list = options2.list && options2.list.map(getRef);
          options2.dict = options2.dict && mapValues(options2.dict, getRef);
        }
        return refs[options.uid];
      }
      Object.assign(schema, options);
      if (typeof schema.callback === "string") {
        try {
          schema.callback = new Function("return " + schema.callback)();
        } catch {
        }
      }
      Object.defineProperty(schema, "uid", { value: globalThis.__schemastery_index__++ });
      Object.setPrototypeOf(schema, Schema2.prototype);
      schema.meta || (schema.meta = {});
      schema.toString = schema.toString.bind(schema);
      return schema;
    }, "Schema");
    Schema2.prototype = Object.create(Function.prototype);
    Schema2.prototype[kSchema] = true;
    Object.defineProperty(Schema2.prototype, "~standard", {
      get() {
        return {
          version: 1,
          vendor: "schemastery",
          validate: __name((value) => {
            try {
              return { value: Schema2.resolve(value, this, {})[0] };
            } catch (error) {
              if (ValidationError.is(error)) {
                return { issues: [{ message: error.message, path: error.options.path }] };
              }
              throw error;
            }
          }, "validate")
        };
      }
    });
    Schema2.ValidationError = ValidationError;
    Schema2.prototype.toJSON = __name(function toJSON() {
      var _a2, _b;
      if (globalThis.__schemastery_refs__) {
        (_a2 = globalThis.__schemastery_refs__)[_b = this.uid] ?? (_a2[_b] = JSON.parse(JSON.stringify({ ...this })));
        return this.uid;
      }
      globalThis.__schemastery_refs__ = { [this.uid]: { ...this } };
      globalThis.__schemastery_refs__[this.uid] = JSON.parse(JSON.stringify({ ...this }));
      const result = { uid: this.uid, refs: globalThis.__schemastery_refs__ };
      globalThis.__schemastery_refs__ = void 0;
      return result;
    }, "toJSON");
    Schema2.prototype.set = __name(function set(key, value) {
      this.dict[key] = value;
      return this;
    }, "set");
    Schema2.prototype.push = __name(function push(value) {
      this.list.push(value);
      return this;
    }, "push");
    function mergeDesc(original, messages) {
      const result = typeof original === "string" ? { "": original } : { ...original };
      for (const locale in messages) {
        const value = messages[locale];
        if ((value == null ? void 0 : value.$description) || (value == null ? void 0 : value.$desc)) {
          result[locale] = value.$description || value.$desc;
        } else if (typeof value === "string") {
          result[locale] = value;
        }
      }
      return result;
    }
    __name(mergeDesc, "mergeDesc");
    function getInner(value) {
      return (value == null ? void 0 : value.$value) ?? (value == null ? void 0 : value.$inner);
    }
    __name(getInner, "getInner");
    function extractKeys(data) {
      return filterKeys(data ?? {}, (key) => !key.startsWith("$"));
    }
    __name(extractKeys, "extractKeys");
    Schema2.prototype.i18n = __name(function i18n(messages) {
      const schema = Schema2(this);
      const desc = mergeDesc(schema.meta.description, messages);
      if (Object.keys(desc).length) schema.meta.description = desc;
      if (schema.dict) {
        schema.dict = mapValues(schema.dict, (inner, key) => {
          return inner.i18n(mapValues(messages, (data) => {
            var _a2;
            return ((_a2 = getInner(data)) == null ? void 0 : _a2[key]) ?? (data == null ? void 0 : data[key]);
          }));
        });
      }
      if (schema.list) {
        schema.list = schema.list.map((inner, index) => {
          return inner.i18n(mapValues(messages, (data = {}) => {
            if (Array.isArray(getInner(data))) return getInner(data)[index];
            if (Array.isArray(data)) return data[index];
            return extractKeys(data);
          }));
        });
      }
      if (schema.inner) {
        schema.inner = schema.inner.i18n(mapValues(messages, (data) => {
          if (getInner(data)) return getInner(data);
          return extractKeys(data);
        }));
      }
      if (schema.sKey) {
        schema.sKey = schema.sKey.i18n(mapValues(messages, (data) => data == null ? void 0 : data.$key));
      }
      return schema;
    }, "i18n");
    Schema2.prototype.extra = __name(function extra(key, value) {
      const schema = Schema2(this);
      schema.meta = { ...schema.meta, [key]: value };
      return schema;
    }, "extra");
    for (const key of ["required", "disabled", "collapse", "hidden", "loose"]) {
      Object.assign(Schema2.prototype, {
        [key](value = true) {
          const schema = Schema2(this);
          schema.meta = { ...schema.meta, [key]: value };
          return schema;
        }
      });
    }
    Schema2.prototype.deprecated = __name(function deprecated() {
      var _a2;
      const schema = Schema2(this);
      (_a2 = schema.meta).badges || (_a2.badges = []);
      schema.meta.badges.push({ text: "deprecated", type: "danger" });
      return schema;
    }, "deprecated");
    Schema2.prototype.experimental = __name(function experimental() {
      var _a2;
      const schema = Schema2(this);
      (_a2 = schema.meta).badges || (_a2.badges = []);
      schema.meta.badges.push({ text: "experimental", type: "warning" });
      return schema;
    }, "experimental");
    Schema2.prototype.pattern = __name(function pattern(regexp) {
      const schema = Schema2(this);
      const pattern2 = pick(regexp, ["source", "flags"]);
      schema.meta = { ...schema.meta, pattern: pattern2 };
      return schema;
    }, "pattern");
    Schema2.prototype.simplify = __name(function simplify(value) {
      if (deepEqual(value, this.meta.default, this.type === "dict")) return null;
      if (isNullable(value)) return value;
      if (this.type === "object" || this.type === "dict") {
        const result = {};
        for (const key in value) {
          const schema = this.type === "object" ? this.dict[key] : this.inner;
          const item = schema == null ? void 0 : schema.simplify(value[key]);
          if (this.type === "dict" || !isNullable(item)) result[key] = item;
        }
        if (deepEqual(result, this.meta.default, this.type === "dict")) return null;
        return result;
      } else if (this.type === "array" || this.type === "tuple") {
        const result = [];
        value.forEach((value2, index) => {
          const schema = this.type === "array" ? this.inner : this.list[index];
          const item = schema ? schema.simplify(value2) : value2;
          result.push(item);
        });
        return result;
      } else if (this.type === "intersect") {
        const result = {};
        for (const item of this.list) {
          Object.assign(result, item.simplify(value));
        }
        return result;
      } else if (this.type === "union") {
        for (const schema of this.list) {
          try {
            Schema2.resolve(value, schema, {});
            return schema.simplify(value);
          } catch {
          }
        }
      }
      return value;
    }, "simplify");
    Schema2.prototype.toString = __name(function toString(inline) {
      var _a2;
      return ((_a2 = formatters[this.type]) == null ? void 0 : _a2.call(formatters, this, inline)) ?? `Schema<${this.type}>`;
    }, "toString");
    Schema2.prototype.role = __name(function role(role, extra) {
      const schema = Schema2(this);
      schema.meta = { ...schema.meta, role, extra };
      return schema;
    }, "role");
    for (const key of ["default", "link", "comment", "description", "max", "min", "step"]) {
      Object.assign(Schema2.prototype, {
        [key](value) {
          const schema = Schema2(this);
          schema.meta = { ...schema.meta, [key]: value };
          return schema;
        }
      });
    }
    var resolvers = {};
    Schema2.extend = __name(function extend(type, resolve) {
      resolvers[type] = resolve;
    }, "extend");
    Schema2.resolve = __name(function resolve(data, schema, options = {}, strict = false) {
      var _a2;
      if (!schema) return [data];
      if ((_a2 = options.ignore) == null ? void 0 : _a2.call(options, data, schema)) return [data];
      if (isNullable(data) && schema.type !== "lazy") {
        if (schema.meta.required) throw new ValidationError(`missing required value`, options);
        let current = schema;
        let fallback = schema.meta.default;
        while ((current == null ? void 0 : current.type) === "intersect" && isNullable(fallback)) {
          current = current.list[0];
          fallback = current == null ? void 0 : current.meta.default;
        }
        if (isNullable(fallback)) return [data];
        data = clone(fallback);
      }
      const callback = resolvers[schema.type];
      if (!callback) throw new ValidationError(`unsupported type "${schema.type}"`, options);
      try {
        return callback(data, schema, options, strict);
      } catch (error) {
        if (!schema.meta.loose) throw error;
        return [schema.meta.default];
      }
    }, "resolve");
    Schema2.from = __name(function from(source) {
      if (isNullable(source)) {
        return Schema2.any();
      } else if (["string", "number", "boolean"].includes(typeof source)) {
        return Schema2.const(source).required();
      } else if (source[kSchema]) {
        return source;
      } else if (typeof source === "function") {
        switch (source) {
          case String:
            return Schema2.string().required();
          case Number:
            return Schema2.number().required();
          case Boolean:
            return Schema2.boolean().required();
          case Function:
            return Schema2.function().required();
          default:
            return Schema2.is(source).required();
        }
      } else {
        throw new TypeError(`cannot infer schema from ${source}`);
      }
    }, "from");
    Schema2.lazy = __name(function lazy(builder) {
      const toJSON = __name(() => {
        if (!schema.inner[kSchema]) {
          schema.inner = schema.builder();
          schema.inner.meta = { ...schema.meta, ...schema.inner.meta };
        }
        return schema.inner.toJSON();
      }, "toJSON");
      const schema = new Schema2({ type: "lazy", builder, inner: { toJSON } });
      return schema;
    }, "lazy");
    Schema2.natural = __name(function natural() {
      return Schema2.number().step(1).min(0);
    }, "natural");
    Schema2.percent = __name(function percent() {
      return Schema2.number().step(0.01).min(0).max(1).role("slider");
    }, "percent");
    Schema2.date = __name(function date() {
      return Schema2.union([
        Schema2.is(Date),
        Schema2.transform(Schema2.string().role("datetime"), (value, options) => {
          const date2 = new Date(value);
          if (isNaN(+date2)) throw new ValidationError(`invalid date "${value}"`, options);
          return date2;
        }, true)
      ]);
    }, "date");
    Schema2.regExp = __name(function regExp(flag = "") {
      return Schema2.union([
        Schema2.is(RegExp),
        Schema2.transform(Schema2.string().role("regexp", { flag }), (value, options) => {
          try {
            return new RegExp(value, flag);
          } catch (e) {
            throw new ValidationError(e.message, options);
          }
        }, true)
      ]);
    }, "regExp");
    Schema2.arrayBuffer = __name(function arrayBuffer(encoding) {
      return Schema2.union([
        Schema2.is(ArrayBuffer),
        Schema2.is(SharedArrayBuffer),
        Schema2.transform(Schema2.any(), (value, options) => {
          if (Binary.isSource(value)) return Binary.fromSource(value);
          throw new ValidationError(`expected ArrayBufferSource but got ${value}`, options);
        }, true),
        ...encoding ? [Schema2.transform(Schema2.string(), (value, options) => {
          try {
            return encoding === "base64" ? Binary.fromBase64(value) : Binary.fromHex(value);
          } catch (e) {
            throw new ValidationError(e.message, options);
          }
        }, true)] : []
      ]);
    }, "arrayBuffer");
    Schema2.extend("lazy", (data, schema, options, strict) => {
      if (!schema.inner[kSchema]) {
        schema.inner = schema.builder();
        schema.inner.meta = { ...schema.meta, ...schema.inner.meta };
      }
      return Schema2.resolve(data, schema.inner, options, strict);
    });
    Schema2.extend("any", (data) => {
      return [data];
    });
    Schema2.extend("never", (data, _, options) => {
      throw new ValidationError(`expected nullable but got ${data}`, options);
    });
    Schema2.extend("const", (data, { value }, options) => {
      if (deepEqual(data, value)) return [value];
      throw new ValidationError(`expected ${value} but got ${data}`, options);
    });
    function checkWithinRange(data, meta, description, options, skipMin = false) {
      const { max = Infinity, min = -Infinity } = meta;
      if (data > max) throw new ValidationError(`expected ${description} <= ${max} but got ${data}`, options);
      if (data < min && !skipMin) throw new ValidationError(`expected ${description} >= ${min} but got ${data}`, options);
    }
    __name(checkWithinRange, "checkWithinRange");
    Schema2.extend("string", (data, { meta }, options) => {
      if (typeof data !== "string") throw new ValidationError(`expected string but got ${data}`, options);
      if (meta.pattern) {
        const regexp = new RegExp(meta.pattern.source, meta.pattern.flags);
        if (!regexp.test(data)) throw new ValidationError(`expect string to match regexp ${regexp}`, options);
      }
      checkWithinRange(data.length, meta, "string length", options);
      return [data];
    });
    function decimalShift(data, digits) {
      const str = data.toString();
      if (str.includes("e")) return data * Math.pow(10, digits);
      const index = str.indexOf(".");
      if (index === -1) return data * Math.pow(10, digits);
      const frac = str.slice(index + 1);
      const integer = str.slice(0, index);
      if (frac.length <= digits) return +(integer + frac.padEnd(digits, "0"));
      return +(integer + frac.slice(0, digits) + "." + frac.slice(digits));
    }
    __name(decimalShift, "decimalShift");
    function isMultipleOf(data, min, step) {
      step = Math.abs(step);
      if (!/^\d+\.\d+$/.test(step.toString())) {
        return (data - min) % step === 0;
      }
      const index = step.toString().indexOf(".");
      const digits = step.toString().slice(index + 1).length;
      return Math.abs(decimalShift(data, digits) - decimalShift(min, digits)) % decimalShift(step, digits) === 0;
    }
    __name(isMultipleOf, "isMultipleOf");
    Schema2.extend("number", (data, { meta }, options) => {
      if (typeof data !== "number") throw new ValidationError(`expected number but got ${data}`, options);
      checkWithinRange(data, meta, "number", options);
      const { step } = meta;
      if (step && !isMultipleOf(data, meta.min ?? 0, step)) {
        throw new ValidationError(`expected number multiple of ${step} but got ${data}`, options);
      }
      return [data];
    });
    Schema2.extend("boolean", (data, _, options) => {
      if (typeof data === "boolean") return [data];
      throw new ValidationError(`expected boolean but got ${data}`, options);
    });
    Schema2.extend("bitset", (data, { bits, meta }, options) => {
      let value = 0, keys = [];
      if (typeof data === "number") {
        value = data;
        for (const key in bits) {
          if (data & bits[key]) {
            keys.push(key);
          }
        }
      } else if (Array.isArray(data)) {
        keys = data;
        for (const key of keys) {
          if (typeof key !== "string") throw new ValidationError(`expected string but got ${key}`, options);
          if (key in bits) value |= bits[key];
        }
      } else {
        throw new ValidationError(`expected number or array but got ${data}`, options);
      }
      if (value === meta.default) return [value];
      return [value, keys];
    });
    Schema2.extend("function", (data, _, options) => {
      if (typeof data === "function") return [data];
      throw new ValidationError(`expected function but got ${data}`, options);
    });
    Schema2.extend("is", (data, { constructor }, options) => {
      var _a2;
      if (typeof constructor === "function") {
        if (data instanceof constructor) return [data];
        throw new ValidationError(`expected ${constructor.name} but got ${data}`, options);
      } else {
        if (isNullable(data)) {
          throw new ValidationError(`expected ${constructor} but got ${data}`, options);
        }
        let prototype = Object.getPrototypeOf(data);
        while (prototype) {
          if (((_a2 = prototype.constructor) == null ? void 0 : _a2.name) === constructor) return [data];
          prototype = Object.getPrototypeOf(prototype);
        }
        throw new ValidationError(`expected ${constructor} but got ${data}`, options);
      }
    });
    function property(data, key, schema, options) {
      try {
        const [value, adapted] = Schema2.resolve(data[key], schema, {
          ...options,
          path: [...options.path || [], key]
        });
        if (adapted !== void 0) data[key] = adapted;
        return value;
      } catch (e) {
        if (!(options == null ? void 0 : options.autofix)) throw e;
        delete data[key];
        return schema.meta.default;
      }
    }
    __name(property, "property");
    Schema2.extend("array", (data, { inner, meta }, options) => {
      if (!Array.isArray(data)) throw new ValidationError(`expected array but got ${data}`, options);
      checkWithinRange(data.length, meta, "array length", options, !isNullable(inner.meta.default));
      return [data.map((_, index) => property(data, index, inner, options))];
    });
    Schema2.extend("dict", (data, { inner, sKey }, options, strict) => {
      if (!isPlainObject(data)) throw new ValidationError(`expected object but got ${data}`, options);
      const result = {};
      for (const key in data) {
        let rKey;
        try {
          rKey = Schema2.resolve(key, sKey, options)[0];
        } catch (error) {
          if (strict) continue;
          throw error;
        }
        result[rKey] = property(data, key, inner, options);
        data[rKey] = data[key];
        if (key !== rKey) delete data[key];
      }
      return [result];
    });
    Schema2.extend("tuple", (data, { list }, options, strict) => {
      if (!Array.isArray(data)) throw new ValidationError(`expected array but got ${data}`, options);
      const result = list.map((inner, index) => property(data, index, inner, options));
      if (strict) return [result];
      result.push(...data.slice(list.length));
      return [result];
    });
    function merge(result, data) {
      for (const key in data) {
        if (key in result) continue;
        result[key] = data[key];
      }
    }
    __name(merge, "merge");
    Schema2.extend("object", (data, { dict }, options, strict) => {
      if (!isPlainObject(data)) throw new ValidationError(`expected object but got ${data}`, options);
      const result = {};
      for (const key in dict) {
        const value = property(data, key, dict[key], options);
        if (!isNullable(value) || key in data) {
          result[key] = value;
        }
      }
      if (!strict) merge(result, data);
      return [result];
    });
    Schema2.extend("union", (data, { list, toString }, options, strict) => {
      const messages = [];
      for (const inner of list) {
        try {
          return Schema2.resolve(data, inner, options, strict);
        } catch (error) {
          messages.push(error);
        }
      }
      throw new ValidationError(`expected ${toString()} but got ${JSON.stringify(data)}`, options);
    });
    Schema2.extend("intersect", (data, { list, toString }, options, strict) => {
      if (!list.length) return [data];
      let result;
      for (const inner of list) {
        const value = Schema2.resolve(data, inner, options, true)[0];
        if (isNullable(value)) continue;
        if (isNullable(result)) {
          result = value;
        } else if (typeof result !== typeof value) {
          throw new ValidationError(`expected ${toString()} but got ${JSON.stringify(data)}`, options);
        } else if (typeof value === "object") {
          merge(result ?? (result = {}), value);
        } else if (result !== value) {
          throw new ValidationError(`expected ${toString()} but got ${JSON.stringify(data)}`, options);
        }
      }
      if (!strict && isPlainObject(data)) merge(result, data);
      return [result];
    });
    Schema2.extend("transform", (data, { inner, callback, preserve }, options) => {
      const [result, adapted = data] = Schema2.resolve(data, inner, options, true);
      if (preserve) {
        return [callback(result)];
      } else {
        return [callback(result), callback(adapted)];
      }
    });
    var formatters = {};
    function defineMethod(name, keys, format) {
      formatters[name] = format;
      Object.assign(Schema2, {
        [name](...args) {
          const schema = new Schema2({ type: name });
          keys.forEach((key, index) => {
            switch (key) {
              case "sKey":
                schema.sKey = args[index] ?? Schema2.string();
                break;
              case "inner":
                schema.inner = Schema2.from(args[index]);
                break;
              case "list":
                schema.list = args[index].map(Schema2.from);
                break;
              case "dict":
                schema.dict = mapValues(args[index], Schema2.from);
                break;
              case "bits": {
                schema.bits = {};
                for (const key2 in args[index]) {
                  if (typeof args[index][key2] !== "number") continue;
                  schema.bits[key2] = args[index][key2];
                }
                break;
              }
              case "callback": {
                const callback = schema.callback = args[index];
                callback["toJSON"] || (callback["toJSON"] = () => callback.toString());
                break;
              }
              case "constructor": {
                const constructor = schema.constructor = args[index];
                if (typeof constructor === "function") {
                  ;
                  constructor["toJSON"] || (constructor["toJSON"] = () => constructor["name"]);
                }
                break;
              }
              default:
                schema[key] = args[index];
            }
          });
          if (name === "object" || name === "dict") {
            schema.meta.default = {};
          } else if (name === "array" || name === "tuple") {
            schema.meta.default = [];
          } else if (name === "bitset") {
            schema.meta.default = 0;
          }
          return schema;
        }
      });
    }
    __name(defineMethod, "defineMethod");
    defineMethod("is", ["constructor"], ({ constructor }) => {
      if (typeof constructor === "function") {
        return constructor.name;
      } else {
        return constructor;
      }
    });
    defineMethod("any", [], () => "any");
    defineMethod("never", [], () => "never");
    defineMethod("const", ["value"], ({ value }) => typeof value === "string" ? JSON.stringify(value) : value);
    defineMethod("string", [], () => "string");
    defineMethod("number", [], () => "number");
    defineMethod("boolean", [], () => "boolean");
    defineMethod("bitset", ["bits"], () => "bitset");
    defineMethod("function", [], () => "function");
    defineMethod("array", ["inner"], ({ inner }) => `${inner.toString(true)}[]`);
    defineMethod("dict", ["inner", "sKey"], ({ inner, sKey }) => `{ [key: ${sKey.toString()}]: ${inner.toString()} }`);
    defineMethod("tuple", ["list"], ({ list }) => `[${list.map((inner) => inner.toString()).join(", ")}]`);
    defineMethod("object", ["dict"], ({ dict }) => {
      if (Object.keys(dict).length === 0) return "{}";
      return `{ ${Object.entries(dict).map(([key, inner]) => {
        return `${key}${inner.meta.required ? "" : "?"}: ${inner.toString()}`;
      }).join(", ")} }`;
    });
    defineMethod("union", ["list"], ({ list }, inline) => {
      const result = list.map(({ toString: format }) => format()).join(" | ");
      return inline ? `(${result})` : result;
    });
    defineMethod("intersect", ["list"], ({ list }) => {
      return `${list.map((inner) => inner.toString(true)).join(" & ")}`;
    });
    defineMethod("transform", ["inner", "callback", "preserve"], ({ inner }, isInner) => inner.toString(isInner));
    module.exports = Schema2;
  }
});
var lib_default = require_index();

// node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/utils.ts
function useI18nText() {
  const composer = useI18n();
  const context = {};
  return (message) => {
    if (!message || typeof message === "string") return message;
    const locales = fallbackWithLocaleChain(context, composer.fallbackLocale.value, composer.locale.value);
    for (const locale of locales) {
      if (locale in message) return message[locale];
    }
    return message[""];
  };
}
var dynamic = ["function", "transform", "is"];
function getChoices(schema) {
  const inner = [];
  const choices = schema.list.filter((item) => {
    if (item.meta.hidden) return;
    if (item.type === "transform") inner.push(item.inner);
    return !dynamic.includes(item.type);
  });
  return choices.length ? choices : inner;
}
function getFallback(schema, required = false) {
  if (!schema || schema.type === "union" && getChoices(schema).length === 1) return;
  return clone(schema.meta.default) ?? (required ? inferFallback(schema) : void 0);
}
function inferFallback(schema) {
  if (schema.type === "string") return "";
  if (schema.type === "number") return 0;
  if (schema.type === "boolean") return false;
  if (["dict", "object", "intersect"].includes(schema.type)) return {};
}
function optional(schema) {
  if (schema.type === "const") return schema;
  if (schema.type === "transform") return optional(schema.inner);
  schema = new lib_default(schema).required(false);
  if (schema.type === "object") {
    schema.dict = mapValues(schema.dict, optional);
  } else if (schema.type === "tuple") {
    schema.list = schema.list.map(optional);
  } else if (schema.type === "intersect") {
    schema.list = schema.list.map(optional);
  } else if (schema.type === "union") {
    schema.list = schema.list.map(optional);
  } else if (schema.type === "dict") {
    schema.inner = optional(schema.inner);
  } else if (schema.type === "array") {
    schema.inner = optional(schema.inner);
  }
  return schema;
}
function useDisabled() {
  const { props } = getCurrentInstance();
  return computed(() => {
    var _a, _b;
    return props.disabled || ((_b = (_a = props.schema) == null ? void 0 : _a.meta) == null ? void 0 : _b.disabled);
  });
}
function useModel(options) {
  let stop;
  const config = ref();
  const { props, emit } = getCurrentInstance();
  const doWatch = () => watch(config, (value) => {
    try {
      if (options == null ? void 0 : options.output) value = options.output(value);
      const schema = optional(lib_default(props.schema));
      if (deepEqual(schema(value), props.schema.meta.default, options == null ? void 0 : options.strict)) value = null;
    } catch {
      return;
    }
    emit("update:modelValue", value);
  }, { deep: true });
  watch(() => [props.modelValue, props.schema], ([value, schema]) => {
    stop == null ? void 0 : stop();
    value ?? (value = getFallback(schema));
    if (options == null ? void 0 : options.input) value = options.input(value);
    config.value = value;
    stop = doWatch();
  }, { deep: true, immediate: true });
  return config;
}
function useEntries() {
  const { props } = getCurrentInstance();
  const entries = useModel({
    strict: true,
    input: (config) => {
      const result = Object.entries(config);
      if (props.schema.type === "array") {
        const padding = (props.schema.meta.min ?? 0) - result.length;
        for (let i = 0; i < padding; i++) {
          result.push(["" + result.length, null]);
        }
      }
      return result;
    },
    output: (config) => {
      if (props.schema.type === "array") {
        return config.map(([, value]) => value);
      }
      const result = {};
      for (const [key, value] of config) {
        if (key in result) throw new Error("duplicate entries");
        result[key] = value;
      }
      return result;
    }
  });
  const isFixedLength = computed(() => {
    return props.schema.meta.min && props.schema.meta.min === props.schema.meta.max;
  });
  const isMax = computed(() => entries.value.length >= props.schema.meta.max);
  const isMin = computed(() => entries.value.length >= props.schema.meta.max);
  const reindex = () => {
    if (props.schema.type !== "array") return;
    for (let i = 0; i < entries.value.length; i++) {
      entries.value[i][0] = "" + i;
    }
  };
  return {
    entries,
    isMax,
    isMin,
    isFixedLength,
    up(index) {
      if (props.schema.type === "dict") {
        entries.value.splice(index - 1, 0, ...entries.value.splice(index, 1));
      } else {
        const temp = entries.value[index][1];
        entries.value[index][1] = entries.value[index - 1][1];
        entries.value[index - 1][1] = temp;
      }
      reindex();
    },
    down(index) {
      if (props.schema.type === "dict") {
        entries.value.splice(index + 1, 0, ...entries.value.splice(index, 1));
      } else {
        const temp = entries.value[index][1];
        entries.value[index][1] = entries.value[index + 1][1];
        entries.value[index + 1][1] = temp;
      }
      reindex();
    },
    del(index) {
      entries.value.splice(index, 1);
      reindex();
    },
    insert(index) {
      entries.value.splice(index, 0, ["", null]);
      reindex();
    }
  };
}
function isConstUnion(schema) {
  return schema.type === "union" && schema.list.every((item) => item.type === "const");
}
function isMultiSelect(schema) {
  if (schema.type === "bitset") return true;
  if (schema.type === "array") return isConstUnion(schema.inner);
}
function isValidColumn(schema) {
  return ["string", "number", "boolean"].includes(schema.type) || isConstUnion(schema) || isMultiSelect(schema);
}
function ensureColumns(entries) {
  entries = entries.filter(([, schema]) => !schema.meta.hidden);
  if (entries.every(([, schema]) => isValidColumn(schema))) return entries;
}
function toColumns(schema) {
  if (isValidColumn(schema)) {
    return [[null, schema]];
  } else if (schema.type === "tuple") {
    return ensureColumns(Object.entries(schema.list));
  } else if (schema.type === "object") {
    return ensureColumns(Object.entries(schema.dict));
  }
}

// node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/index.ts
import SchemaBase from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/base.vue";
import Primitive from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/primitive.vue";
import SchemaCheckbox from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/extensions/checkbox.vue";
import SchemaGroup from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/extensions/group.vue";
import SchemaIntersect from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/extensions/intersect.vue";
import SchemaObject from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/extensions/object.vue";
import SchemaRadio from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/extensions/radio.vue";
import SchemaMultiSelect from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/extensions/multiselect.vue";
import SchemaTable from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/extensions/table.vue";
import SchemaTextarea from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/extensions/textarea.vue";
import SchemaTuple from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/extensions/tuple.vue";
import SchemaUnion from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/extensions/union.vue";
import KBadge from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/badge.vue";
import KSchema from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/schema.vue";
import KForm from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/form.vue";
import "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/styles/index.scss";

// node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/index.ts
import IconAdd from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/add.vue";
import IconArrowDown from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/arrow-down.vue";
import IconArrowUp from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/arrow-up.vue";
import IconBranch from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/branch.vue";
import IconClose from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/close.vue";
import IconCode from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/code.vue";
import IconCollapse from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/collapse.vue";
import IconDelete from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/delete.vue";
import IconEllipsis from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/ellipsis.vue";
import IconExpand from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/expand.vue";
import IconExternal from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/external.vue";
import IconEyeSlash from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/eye-slash.vue";
import IconEye from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/eye.vue";
import IconInsertAfter from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/insert-after.vue";
import IconInsertBefore from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/insert-before.vue";
import IconInvalid from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/invalid.vue";
import IconRedo from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/redo.vue";
import IconReset from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/reset.vue";
import IconSquareCheck from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/square-check.vue";
import IconSquareEmpty from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/square-empty.vue";
import IconUndo from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/icons/undo.vue";

// node_modules/.bun/schemastery-vue@7.3.15+d2adbfbcd7c7907c/node_modules/schemastery-vue/src/index.ts
var extensions = /* @__PURE__ */ new Set();
var form = Object.assign(SchemaBase, {
  Form: KForm,
  Badge: KBadge,
  Schema: KSchema,
  useModel,
  useEntries,
  useDisabled,
  getFallback,
  extensions,
  install(app) {
    app.provide("__SCHEMASTERY_EXTENSIONS__", extensions);
    app.component("k-form", KForm);
    app.component("k-badge", KBadge);
    app.component("k-schema", KSchema);
  }
});
form.extensions.add({
  type: "bitset",
  role: "select",
  component: SchemaMultiSelect,
  validate: (value) => typeof value === "number" || Array.isArray(value) && value.every((v) => typeof v === "string")
});
form.extensions.add({
  type: "array",
  role: "select",
  component: SchemaMultiSelect,
  validate: (value) => Array.isArray(value) && value.every((v) => typeof v === "string")
});
form.extensions.add({
  type: "bitset",
  component: SchemaCheckbox,
  validate: (value) => typeof value === "number" || Array.isArray(value) && value.every((v) => typeof v === "string")
});
form.extensions.add({
  type: "array",
  role: "checkbox",
  component: SchemaCheckbox,
  validate: (value) => Array.isArray(value) && value.every((v) => typeof v === "string")
});
form.extensions.add({
  type: "array",
  component: SchemaGroup,
  validate: (value) => Array.isArray(value)
});
form.extensions.add({
  type: "dict",
  component: SchemaGroup,
  validate: (value) => typeof value === "object"
});
form.extensions.add({
  type: "object",
  component: SchemaObject,
  validate: (value) => typeof value === "object"
});
form.extensions.add({
  type: "intersect",
  component: SchemaIntersect,
  validate: (value) => typeof value === "object"
});
form.extensions.add({
  type: "union",
  role: "radio",
  component: SchemaRadio
});
form.extensions.add({
  type: "array",
  role: "table",
  component: SchemaTable,
  validate: (value, schema) => Array.isArray(value) && !!toColumns(schema.inner)
});
form.extensions.add({
  type: "dict",
  role: "table",
  component: SchemaTable,
  validate: (value, schema) => typeof value === "object" && !!toColumns(schema.inner)
});
form.extensions.add({
  type: "string",
  role: "textarea",
  component: SchemaTextarea,
  validate: (value) => typeof value === "string"
});
form.extensions.add({
  type: "tuple",
  component: SchemaTuple,
  validate: (value) => Array.isArray(value)
});
form.extensions.add({
  type: "union",
  component: SchemaUnion
});
var src_default = form;

// node_modules/.bun/@koishijs+components@1.5.22+d2adbfbcd7c7907c/node_modules/@koishijs/components/client/form/index.ts
import Computed from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+components@1.5.22+d2adbfbcd7c7907c/node_modules/@koishijs/components/client/form/computed.vue";
import Filter from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+components@1.5.22+d2adbfbcd7c7907c/node_modules/@koishijs/components/client/form/k-filter.vue";
src_default.extensions.add({
  type: "union",
  role: "computed",
  component: Computed
});
function form_default(app) {
  app.use(src_default);
  app.component("k-filter", Filter);
}

// node_modules/.bun/@koishijs+components@1.5.22+d2adbfbcd7c7907c/node_modules/@koishijs/components/client/virtual/index.ts
import VirtualList from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+components@1.5.22+d2adbfbcd7c7907c/node_modules/@koishijs/components/client/virtual/list.vue";
function virtual_default(app) {
  app.component("virtual-list", VirtualList);
}

// node_modules/.bun/@koishijs+components@1.5.22+d2adbfbcd7c7907c/node_modules/@koishijs/components/client/index.ts
import Comment from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+components@1.5.22+d2adbfbcd7c7907c/node_modules/@koishijs/components/client/k-comment.vue";
import ImageViewer from "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+components@1.5.22+d2adbfbcd7c7907c/node_modules/@koishijs/components/client/image-viewer.vue";
import "D:/Python_Project/Moebot-NEXT/node_modules/.bun/@koishijs+components@1.5.22+d2adbfbcd7c7907c/node_modules/@koishijs/components/client/index.scss";
function client_default(app) {
  app.use(form_default);
  app.use(virtual_default);
  app.component("k-comment", Comment);
  app.component("k-image-viewer", ImageViewer);
}

export {
  noop,
  isNullable,
  isNonNullable,
  isPlainObject,
  filterKeys,
  mapValues,
  pick,
  omit,
  defineProperty,
  contain,
  intersection,
  difference,
  union,
  deduplicate,
  remove,
  makeArray,
  is,
  Binary,
  base64ToArrayBuffer,
  arrayBufferToBase64,
  hexToArrayBuffer,
  arrayBufferToHex,
  clone,
  deepEqual,
  capitalize,
  uncapitalize,
  camelCase,
  paramCase,
  snakeCase,
  camelize,
  hyphenate,
  formatProperty,
  trimSlash,
  sanitize,
  Time,
  lib_default,
  useI18nText,
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
  IconEyeSlash,
  IconEye,
  IconInsertAfter,
  IconInsertBefore,
  IconInvalid,
  IconRedo,
  IconReset,
  IconSquareCheck,
  IconSquareEmpty,
  IconUndo,
  Primitive,
  form,
  src_default,
  VirtualList,
  client_default
};
//# sourceMappingURL=chunk-BZ4TRKRY.js.map
