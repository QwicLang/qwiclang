# Heterogeneous Collections & Literals Implementation Guide

## Overview

This document outlines a **v0-compatible** approach to add heterogeneous array, dictionary, and tuple literals to QwicLang without requiring full generics.

### Goal Syntax

```qwic
public func main() {
    // Heterogeneous array literal
    let a = [1, 2, "a", "b"]
    
    // Dictionary literal
    let b = {"a": 1, "b": 2}
    
    // Tuple literal (N-ary)
    let c = (1, 2, 3, 4)
    
    // Iteration
    for item in a {
        print(item)
    }
    
    for key in b.keys() {
        print(f"{key}: {b[key]}")
    }
}
```

---

## Design Decisions

### 1. **Tagged Values at Runtime**

Instead of strings-only, we'll use **tagged values** in C:

```c
typedef enum {
    QWIC_TYPE_INT,
    QWIC_TYPE_FLOAT,
    QWIC_TYPE_STRING,
    QWIC_TYPE_BOOL,
} qwic_type;

typedef struct {
    qwic_type type;
    union {
        int64_t i;
        double f;
        char *s;
        bool b;
    } value;
} qwic_value;
```

### 2. **Array Literals** → Runtime `list<any>`

- Syntax: `[1, 2, "a", "b"]`
- Type at compile time: `list` (no element type tracking)
- Runtime storage: `qwic_value[]` (tagged values)

### 3. **Dictionary Literals** → Runtime `dictionary<any>`

- Syntax: `{"a": 1, "b": 2}`
- Type at compile time: `dictionary`
- Runtime: keys are strings, values are `qwic_value`

### 4. **Tuple Literals** → Runtime N-ary tuples

- Syntax: `(1, 2, 3, 4)`
- Type at compile time: `tuple` (no arity tracking yet)
- Runtime: flexible-length `qwic_value[]`

### 5. **For Loops** (Early Addition)

- Syntax: `for item in collection { ... }`
- Supports: lists, dictionaries, tuples

---

## Implementation Phases

### Phase A: Lexer (No Changes Needed)
The lexer already recognizes `[`, `]`, `{`, `}`, `:`, `,`. ✅

---

### Phase B: Parser Updates

**Files to modify:**
- `internal/token/token.go` — add `For`, `In`, `LBracket`, `RBracket` keywords (already exist)
- `internal/ast/ast.go` — add AST nodes
- `internal/parser/parser.go` — add parsing logic

**AST Nodes to add:**

```go
// Array literal: [1, 2, "a", "b"]
type ArrayLiteralExpression struct {
    Elements []Expression
    Pos      token.Position
}

func (*ArrayLiteralExpression) expressionNode() {}

// Dictionary literal: {"a": 1, "b": 2}
type DictionaryLiteralExpression struct {
    Pairs []struct {
        Key   Expression  // Must be string
        Value Expression
    }
    Pos token.Position
}

func (*DictionaryLiteralExpression) expressionNode() {}

// Tuple literal: (1, 2, 3)
type TupleLiteralExpression struct {
    Elements []Expression
    Pos      token.Position
}

func (*TupleLiteralExpression) expressionNode() {}

// Index access: a[0]
type IndexExpression struct {
    Left  Expression
    Index Expression
    Pos   token.Position
}

func (*IndexExpression) expressionNode() {}

// For loop: for item in collection { ... }
type ForStatement struct {
    Variable   string
    Iterable   Expression
    Body       *BlockStatement
    Pos        token.Position
}

func (*ForStatement) statementNode() {}
```

**Parser changes:**

```go
// In parseCallExpression, after dot handling, add index access:
if parser.match(token.LBracket) {
    start := parser.previous().Start
    index := parser.parseExpression()
    parser.consume(token.RBracket, "expected ']' after index")
    expression = &ast.IndexExpression{
        Left:  expression,
        Index: index,
        Pos:   start,
    }
    continue
}

// In parsePrimaryExpression, handle array/dict/tuple literals:
case parser.match(token.LBracket):
    return parser.parseArrayLiteral(current.Start)
case parser.match(token.LBrace):
    return parser.parseDictionaryLiteral(current.Start)
case parser.match(token.LParen):
    // This is tricky: distinguish (expr) from (1, 2) tuple
    // Try to parse as tuple first
    return parser.parseTupleLiteral(current.Start)

// In parseStatement, add:
case parser.match(token.For):
    return parser.parseForStatement(parser.previous().Start)
```

---

### Phase C: Type System Updates

**File: `internal/types/types.go`**

No changes needed. We already have `List`, `Dictionary`, `Tuple` kinds.

**However, add to the Type struct:**

```go
type Type struct {
    Kind       Kind
    ElementType *Type  // For List<T> (future)
}
```

For now, `ElementType` is always `nil` (homogeneous at compile time).

---

### Phase D: Semantic Analysis

**File: `internal/sema/sema.go`**

```go
// Validate array literals
case *ast.ArrayLiteralExpression:
    for _, elem := range node.Elements {
        s.analyzeExpression(elem)
        // All elements allowed, type checked at runtime
    }
    return types.ListType

// Validate dictionary literals
case *ast.DictionaryLiteralExpression:
    for _, pair := range node.Pairs {
        keyType := s.analyzeExpression(pair.Key)
        if keyType.Kind != types.String {
            s.reportError(pair.Key.Position(), "dictionary keys must be strings")
        }
        s.analyzeExpression(pair.Value)
    }
    return types.DictType

// Validate tuple literals
case *ast.TupleLiteralExpression:
    for _, elem := range node.Elements {
        s.analyzeExpression(elem)
    }
    return types.TupleType

// Index access: a[0]
case *ast.IndexExpression:
    leftType := s.analyzeExpression(node.Left)
    s.analyzeExpression(node.Index)
    
    switch leftType.Kind {
    case types.List, types.Dictionary, types.Tuple:
        return types.Any  // We don't know exact type at compile time
    default:
        s.reportError(node.Position(), "cannot index non-collection type")
        return types.InvalidType
    }

// For loop
case *ast.ForStatement:
    iterType := s.analyzeExpression(node.Iterable)
    switch iterType.Kind {
    case types.List, types.Dictionary, types.Tuple:
        // OK
    default:
        s.reportError(node.Iterable.Position(), "can only iterate over collections")
    }
    
    // Register loop variable
    s.scope.Set(node.Variable, types.Any)
    s.analyzeBlockStatement(node.Body)
```

---

### Phase E: IR Generation

**File: `internal/ir/ir.go`**

Add IR instructions:

```go
// Array construction
type ArrayConstruction struct {
    Target   string
    Elements []string  // Variable names or constants
    Type     types.Type
}

// Dictionary construction
type DictionaryConstruction struct {
    Target string
    Pairs  []struct {
        Key   string
        Value string
    }
    Type types.Type
}

// Tuple construction
type TupleConstruction struct {
    Target   string
    Elements []string
    Type     types.Type
}

// Index access
type IndexAccess struct {
    Target    string
    Container string
    Index     string
    Type      types.Type
}

// For loop (requires label-based IR)
type IterationStart struct {
    Variable  string
    Container string
    LoopLabel string
}

type IterationNext struct {
    LoopLabel   string
    ContinueLabel string
    BreakLabel  string
}
```

**IR generation in `internal/sema` during lowering:**

```go
// For array literal [1, 2, "a"]
// Generate:
// %0 = array_new()
// array_push(%0, 1)
// array_push(%0, 2)
// array_push(%0, "a")
```

---

### Phase F: C Code Generation

**File: `internal/codegen/c.go` and `runtime/qwic_runtime.c`**

#### New runtime functions:

```c
// Tagged value creation
qwic_value qwic_value_int(int64_t i);
qwic_value qwic_value_float(double f);
qwic_value qwic_value_string(const char *s);
qwic_value qwic_value_bool(bool b);

// Array (list) operations with tagged values
void *qwic_list_new_any(void);
void qwic_list_push_any(void *list, qwic_value value);
qwic_value qwic_list_get_any(void *list, int64_t index);
qwic_value *qwic_list_iterator_new(void *list);

// Dictionary operations
void *qwic_dict_new_any(void);
void qwic_dict_set_any(void *dict, const char *key, qwic_value value);
qwic_value qwic_dict_get_any(void *dict, const char *key);
const char **qwic_dict_keys(void *dict);

// Tuple operations
void *qwic_tuple_new_any(qwic_value *elements, int64_t count);
qwic_value qwic_tuple_get_any(void *tuple, int64_t index);

// Value printing
void qwic_print_value(qwic_value v);

// For iteration
typedef struct {
    void *collection;
    int64_t index;
    int64_t length;
    int type; // 0=list, 1=dict, 2=tuple
} qwic_iterator;

qwic_iterator qwic_iterator_new(void *collection, int type);
bool qwic_iterator_next(qwic_iterator *it, qwic_value *out);
```

#### C generation for literals:

```c
// For: let a = [1, 2, "a", "b"]
void *qw_tmp_a = qwic_list_new_any();
qwic_list_push_any(qw_tmp_a, qwic_value_int(1));
qwic_list_push_any(qw_tmp_a, qwic_value_int(2));
qwic_list_push_any(qw_tmp_a, qwic_value_string("a"));
qwic_list_push_any(qw_tmp_a, qwic_value_string("b"));
```

#### C generation for for loops:

```c
// for item in a { print(item) }
qwic_iterator qw_iter = qwic_iterator_new(qw_tmp_a, 0);  // 0 = list
qwic_value qw_item;
while (qwic_iterator_next(&qw_iter, &qw_item)) {
    qwic_print_value(qw_item);
}
```

---

## Implementation Files

### 1. `internal/token/token.go` (Already has needed tokens)

No changes needed. `LBracket`, `RBracket`, `Colon` already exist.
Add: `For`, `In` keywords if not present.

---

### 2. `internal/ast/ast.go` (Add to file)

```go
type ArrayLiteralExpression struct {
    Elements []Expression
    Pos      token.Position
}

func (*ArrayLiteralExpression) expressionNode() {}
func (e *ArrayLiteralExpression) Position() token.Position { return e.Pos }

type DictionaryLiteralExpression struct {
    Pairs []DictionaryPair
    Pos   token.Position
}

type DictionaryPair struct {
    Key   Expression
    Value Expression
}

func (*DictionaryLiteralExpression) expressionNode() {}
func (e *DictionaryLiteralExpression) Position() token.Position { return e.Pos }

type TupleLiteralExpression struct {
    Elements []Expression
    Pos      token.Position
}

func (*TupleLiteralExpression) expressionNode() {}
func (e *TupleLiteralExpression) Position() token.Position { return e.Pos }

type IndexExpression struct {
    Left  Expression
    Index Expression
    Pos   token.Position
}

func (*IndexExpression) expressionNode() {}
func (e *IndexExpression) Position() token.Position { return e.Pos }

type ForStatement struct {
    Variable string
    Iterable Expression
    Body     *BlockStatement
    Pos      token.Position
}

func (*ForStatement) statementNode() {}
func (s *ForStatement) Position() token.Position { return s.Pos }
```

---

### 3. `internal/parser/parser.go` (Add methods)

```go
func (parser *Parser) parseArrayLiteral(start token.Position) ast.Expression {
    var elements []ast.Expression
    
    if !parser.check(token.RBracket) {
        for {
            elements = append(elements, parser.parseExpression())
            if !parser.match(token.Comma) {
                break
            }
        }
    }
    
    parser.consume(token.RBracket, "expected ']' after array elements")
    return &ast.ArrayLiteralExpression{Elements: elements, Pos: start}
}

func (parser *Parser) parseDictionaryLiteral(start token.Position) ast.Expression {
    var pairs []ast.DictionaryPair
    
    if !parser.check(token.RBrace) {
        for {
            key := parser.parseExpression()
            parser.consume(token.Colon, "expected ':' after dictionary key")
            value := parser.parseExpression()
            pairs = append(pairs, ast.DictionaryPair{Key: key, Value: value})
            
            if !parser.match(token.Comma) {
                break
            }
        }
    }
    
    parser.consume(token.RBrace, "expected '}' after dictionary pairs")
    return &ast.DictionaryLiteralExpression{Pairs: pairs, Pos: start}
}

func (parser *Parser) parseTupleLiteral(start token.Position) ast.Expression {
    var elements []ast.Expression
    
    if !parser.check(token.RParen) {
        elements = append(elements, parser.parseExpression())
        
        // If only one element and no comma, it's a grouped expression
        if parser.check(token.RParen) {
            parser.consume(token.RParen, "expected ')' after expression")
            return elements[0]
        }
        
        // Multiple elements = tuple
        parser.consume(token.Comma, "expected ',' in tuple")
        elements = append(elements, parser.parseExpression())
        
        for parser.match(token.Comma) {
            if parser.check(token.RParen) {
                break
            }
            elements = append(elements, parser.parseExpression())
        }
    }
    
    parser.consume(token.RParen, "expected ')' after tuple elements")
    return &ast.TupleLiteralExpression{Elements: elements, Pos: start}
}

func (parser *Parser) parseForStatement(start token.Position) ast.Statement {
    variable, ok := parser.consumeIdentifier("expected variable name in for loop")
    if !ok {
        parser.synchronizeStatement()
        return nil
    }
    
    if !parser.match(token.In) {
        // Create In token if needed
        parser.errorAtCurrent("expected 'in' in for loop")
        parser.synchronizeStatement()
        return nil
    }
    
    iterable := parser.parseExpression()
    body := parser.parseBlock()
    if body == nil {
        parser.synchronizeStatement()
        return nil
    }
    
    return &ast.ForStatement{
        Variable: variable.Lexeme,
        Iterable: iterable,
        Body:     body,
        Pos:      start,
    }
}

// In parseCallExpression, add index access:
func (parser *Parser) parseCallExpression() ast.Expression {
    expression := parser.parsePrimaryExpression()

    for {
        if parser.match(token.LBracket) {
            start := parser.previous().Start
            index := parser.parseExpression()
            parser.consume(token.RBracket, "expected ']' after index")
            expression = &ast.IndexExpression{
                Left:  expression,
                Index: index,
                Pos:   start,
            }
            continue
        }
        // ... rest of dot and call handling
    }
}

// In parsePrimaryExpression, add:
case parser.match(token.LBracket):
    return parser.parseArrayLiteral(current.Start)
case parser.match(token.LBrace):
    return parser.parseDictionaryLiteral(current.Start)
// Modify paren handling
case parser.match(token.LParen):
    return parser.parseTupleLiteral(current.Start)
```

---

### 4. `internal/token/token.go` (Ensure In/For exist)

```go
const (
    // ... existing tokens
    For  Kind = iota  // if not present
    In   Kind = iota  // if not present
)

func LookupIdentifier(ident string) Kind {
    keywords := map[string]Kind{
        // ... existing
        "for": For,
        "in":  In,
    }
    // ...
}
```

---

### 5. `runtime/qwic_runtime.h` (Add signatures)

```c
// Tagged value type
typedef enum {
    QWIC_TYPE_INT,
    QWIC_TYPE_FLOAT,
    QWIC_TYPE_STRING,
    QWIC_TYPE_BOOL,
    QWIC_TYPE_NULL,
} qwic_value_type;

typedef struct {
    qwic_value_type type;
    union {
        int64_t i;
        double f;
        char *s;
        bool b;
    } value;
} qwic_value;

// Value creation
qwic_value qwic_value_int(int64_t i);
qwic_value qwic_value_float(double f);
qwic_value qwic_value_string(const char *s);
qwic_value qwic_value_bool(bool b);
qwic_value qwic_value_null(void);

// Array operations
void *qwic_list_new_any(void);
void qwic_list_push_any(void *list, qwic_value value);
qwic_value qwic_list_get_any(void *list, int64_t index);
int64_t qwic_list_length_any(void *list);

// Dictionary operations
void *qwic_dict_new_any(void);
void qwic_dict_set_any(void *dict, const char *key, qwic_value value);
qwic_value qwic_dict_get_any(void *dict, const char *key);
const char **qwic_dict_keys(void *dict);
int64_t qwic_dict_length_any(void *dict);

// Tuple operations
void *qwic_tuple_new_any(const qwic_value *elements, int64_t count);
qwic_value qwic_tuple_get_any(void *tuple, int64_t index);
int64_t qwic_tuple_length_any(void *tuple);

// Printing
void qwic_print_value(qwic_value v);

// Iteration
typedef struct {
    void *collection;
    int64_t index;
    int64_t length;
    int type;
    const char **dict_keys;
    int key_index;
} qwic_iterator;

qwic_iterator qwic_iterator_new(void *collection, int type);
bool qwic_iterator_next(qwic_iterator *it, qwic_value *out);
```

---

### 6. `runtime/qwic_runtime.c` (Implementation)

```c
#include "qwic_runtime.h"
#include <stdlib.h>
#include <string.h>

// Existing list, dict, tuple structs stay the same

// Value creation
qwic_value qwic_value_int(int64_t i) {
    return (qwic_value){ .type = QWIC_TYPE_INT, .value = {.i = i} };
}

qwic_value qwic_value_float(double f) {
    return (qwic_value){ .type = QWIC_TYPE_FLOAT, .value = {.f = f} };
}

qwic_value qwic_value_string(const char *s) {
    return (qwic_value){ .type = QWIC_TYPE_STRING, .value = {.s = qwic_copy_string(s)} };
}

qwic_value qwic_value_bool(bool b) {
    return (qwic_value){ .type = QWIC_TYPE_BOOL, .value = {.b = b} };
}

qwic_value qwic_value_null(void) {
    return (qwic_value){ .type = QWIC_TYPE_NULL, .value = {.i = 0} };
}

// Print a tagged value
void qwic_print_value(qwic_value v) {
    switch (v.type) {
        case QWIC_TYPE_INT:
            qwic_print_int(v.value.i);
            break;
        case QWIC_TYPE_FLOAT:
            qwic_print_float(v.value.f);
            break;
        case QWIC_TYPE_STRING:
            qwic_print_string(v.value.s);
            break;
        case QWIC_TYPE_BOOL:
            qwic_print_bool(v.value.b);
            break;
        case QWIC_TYPE_NULL:
            printf("null\n");
            break;
    }
}

// List operations (wrapper over existing qwic_list)
typedef struct {
    size_t length;
    size_t capacity;
    qwic_value *items;
} qwic_list_any;

void *qwic_list_new_any(void) {
    qwic_list_any *list = qwic_alloc(sizeof(qwic_list_any));
    list->length = 0;
    list->capacity = 0;
    list->items = NULL;
    return list;
}

void qwic_list_push_any(void *list, qwic_value value) {
    if (list == NULL) return;
    qwic_list_any *typed = (qwic_list_any *)list;
    
    if (typed->capacity <= typed->length) {
        size_t new_cap = typed->capacity == 0 ? 4 : typed->capacity * 2;
        qwic_value *new_items = qwic_alloc(sizeof(qwic_value) * new_cap);
        if (typed->items) {
            memcpy(new_items, typed->items, sizeof(qwic_value) * typed->length);
            qwic_free(typed->items);
        }
        typed->items = new_items;
        typed->capacity = new_cap;
    }
    
    typed->items[typed->length++] = value;
}

qwic_value qwic_list_get_any(void *list, int64_t index) {
    if (list == NULL || index < 0) {
        return qwic_value_null();
    }
    qwic_list_any *typed = (qwic_list_any *)list;
    if ((size_t)index >= typed->length) {
        return qwic_value_null();
    }
    return typed->items[index];
}

int64_t qwic_list_length_any(void *list) {
    if (list == NULL) return 0;
    return (int64_t)((qwic_list_any *)list)->length;
}

// ... Similar for dict_new_any, dict_set_any, dict_get_any, dict_keys

// Tuple operations
typedef struct {
    int64_t length;
    qwic_value *items;
} qwic_tuple_any;

void *qwic_tuple_new_any(const qwic_value *elements, int64_t count) {
    qwic_tuple_any *tuple = qwic_alloc(sizeof(qwic_tuple_any));
    tuple->length = count;
    tuple->items = qwic_alloc(sizeof(qwic_value) * count);
    memcpy(tuple->items, elements, sizeof(qwic_value) * count);
    return tuple;
}

qwic_value qwic_tuple_get_any(void *tuple, int64_t index) {
    if (tuple == NULL || index < 0) {
        return qwic_value_null();
    }
    qwic_tuple_any *typed = (qwic_tuple_any *)tuple;
    if (index >= typed->length) {
        return qwic_value_null();
    }
    return typed->items[index];
}

int64_t qwic_tuple_length_any(void *tuple) {
    if (tuple == NULL) return 0;
    return ((qwic_tuple_any *)tuple)->length;
}

// Iteration
qwic_iterator qwic_iterator_new(void *collection, int type) {
    qwic_iterator it = {
        .collection = collection,
        .index = 0,
        .type = type,
        .dict_keys = NULL,
        .key_index = 0,
    };
    
    if (type == 0) {  // list
        it.length = qwic_list_length_any(collection);
    } else if (type == 1) {  // dict
        it.length = qwic_dict_length_any(collection);
        it.dict_keys = qwic_dict_keys(collection);
    } else if (type == 2) {  // tuple
        it.length = qwic_tuple_length_any(collection);
    }
    
    return it;
}

bool qwic_iterator_next(qwic_iterator *it, qwic_value *out) {
    if (it->type == 0) {  // list
        if (it->index >= it->length) return false;
        *out = qwic_list_get_any(it->collection, it->index++);
        return true;
    } else if (it->type == 1) {  // dict - iterate over values
        if (it->key_index >= it->length) return false;
        const char *key = it->dict_keys[it->key_index++];
        *out = qwic_dict_get_any(it->collection, key);
        return true;
    } else if (it->type == 2) {  // tuple
        if (it->index >= it->length) return false;
        *out = qwic_tuple_get_any(it->collection, it->index++);
        return true;
    }
    return false;
}
```

---

## Testing

Create tests in `tests/integration/`:

```qwic
// test_heterogeneous_arrays.qw
public func main() {
    let a = [1, 2, "hello", "world"]
    print(a[0])  // 1
    print(a[2])  // "hello"
    print(a.length())  // 4
}

// test_dictionaries.qw
public func main() {
    let b = {"name": "Alice", "age": "30"}
    print(b["name"])  // "Alice"
}

// test_for_loop.qw
public func main() {
    let items = [1, 2, 3]
    for item in items {
        print(item)
    }
}

// test_tuples.qw
public func main() {
    let t = (1, 2, 3)
    print(t[0])  // 1
    print(t.length())  // 3
}
```

---

## Priority Order

1. **Lexer** — Already done ✅
2. **Parser** — Add AST nodes and parsing (1-2 days)
3. **Type system** — Minimal changes (1 day)
4. **Semantic analyzer** — Validate literals (1 day)
5. **IR** — Generate allocation/push instructions (2 days)
6. **Runtime** — Implement `qwic_value`, array/dict/tuple ops (3 days)
7. **Codegen** — Generate C with tagged values (2 days)
8. **Testing** — Integration tests (1 day)

**Total: ~2 weeks** (assuming full-time)

---

## Performance Considerations

- **Tagging cost**: +8 bytes per value (type enum + padding)
- **No overhead** on operations — dispatch at runtime, not compile time
- **Memory**: Minimal — reuses existing allocation strategy
- **Speed**: Same as before — linear storage, O(1) indexing

This approach is **faster than generics monomorphization** (no code duplication) and **simpler than `any` containers** (no reference counting overhead).
