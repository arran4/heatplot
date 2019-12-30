%{
package main;

import __yyfmt__ "fmt"

var yyResult *Function
%}

%token<expr> Highest
%token<float> FLOAT
%token<s> VAR
%type<expr> expr
%type<equals> equals

%union {
    equals *Equals
    float float64
    s string
    expr Expression
 }

%right '='
%left '+' '-'
%left '*' '/' '%' '^'
%right Highest

%%
input
    : equals { yyResult = &Function{ Equals: $1 } }
    ;

equals
    : expr '=' expr { $$ = &Equals { LHS: $1, RHS: $3 } }
    ;

expr: FLOAT             { $$ = &Const{Value: $1} }
    | VAR               { $$ = &Var{ Var: $1 } }
    | expr '+' expr     { $$ = &Plus{ LHS: $1, RHS: $3, } }
    | expr '-' expr     { $$ = &Subtract{ LHS: $1, RHS: $3, } }
    | expr '*' expr     { $$ = &Multiply{ LHS: $1, RHS: $3, } }
    | expr '/' expr     { $$ = &Divide{ LHS: $1, RHS: $3, } }
    // | expr '%' expr     { $$ = &Modulus{ LHS: $1, RHS: $3, } }
    | expr '^' expr     { $$ = &Power{ LHS: $1, RHS: $3, } }
    | '+' expr  %prec Highest    { $$ = $2 }
    | '-' expr  %prec Highest    { $$ = &Negate{ Expr: $2 } }
    | '(' expr ')'              { $$ = &Brackets{ Expr: $2 } }
    ;

%%
