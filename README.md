## Changelog
* 05/15/2022:
    - Finished Chapter 6: Parsing Expressions  
    - Bug: Tokens with TokenType Number have integer Literal values.
      These should be floating point numbers, given that Lox only has one numerical type.
    - Need to verify if AstPrinter is rendering correctly or if there's a parsing bug
    - Verify that errors act as they should. Using panic() and recover() instead of
      propagating errors up may not be the best way.
* 05/16/2022:
    - Halfway through Chapter 7: Evaluating Expressions
    - Fixed bug reg: TokenType Number
    - Probably easier to verify functionality if the Interpreter is wired up