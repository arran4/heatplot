%{
package main;
%}

%%
input
    : "h" { *expression = $1; }
    ;

%%
