%{
package heatPlot;

import __yyfmt__ "fmt"

var yyResult *Function
%}

%token<expr> Highest
%token<float> FLOAT
%token<s> VAR FUNCNAME
%type<expr> expr

%union {
    float float64
    s string
    expr Expression
 }

%right '='
%left '+' '-'
%left '*' '/' '%' '^' ','
%right Highest FUNCNAME

%%
input
    : expr '=' expr { yyResult = &Function{ Equals: &Equals { LHS: $1, RHS: $3 } } }
    ;

expr: FLOAT             { $$ = &Const{Value: $1} }
    | VAR               { $$ = &Var{ Var: $1 } }
    | expr '+' expr     { $$ = &Plus{ LHS: $1, RHS: $3, } }
    | expr '-' expr     { $$ = &Subtract{ LHS: $1, RHS: $3, } }
    | expr '*' expr     { $$ = &Multiply{ LHS: $1, RHS: $3, } }
    | expr '/' expr     { $$ = &Divide{ LHS: $1, RHS: $3, } }
    | expr '%' expr     { $$ = &Modulus{ LHS: $1, RHS: $3, } }
    | expr '^' expr     { $$ = &Power{ LHS: $1, RHS: $3, } }
    | '+' expr  %prec Highest    { $$ = $2 }
    | '-' expr  %prec Highest    { $$ = &Negate{ Expr: $2 } }
    | expr FUNCNAME expr     { $$ = &DoubleFunction{ Infix: true, Name: $2, Expr1: $1, Expr2: $3, } }
    | FUNCNAME '(' expr ')'  { $$ = &SingleFunction{ Name: $1, Expr: $3 } }
    | FUNCNAME '(' expr ',' expr ')' { $$ = &DoubleFunction{ Infix: false, Name: $1, Expr1: $3, Expr2: $5 } }
    | '(' expr ')'            { $$ = &Brackets{ Expr: $2 } }
    ;

%%
