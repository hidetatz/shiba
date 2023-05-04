#pragma once

#include <cstdint>
#include <string>

namespace shiba {

/*
 * common
 */

typedef uint8_t u8;
typedef uint16_t u16;
typedef uint32_t u32;
typedef uint64_t u64;

typedef int8_t i8;
typedef int16_t i16;
typedef int32_t i32;
typedef int64_t i64;

/*
 * token
 */

enum class TokenType {
	illegal,
	eof,
	ident,

	// values
	_int,
	_float
	string,
	
	// panctuators
	assign,
	plus,
	minus,
	bang,
	star,
	slash,
	eq,
	neq,
	lt,
	gt,
	comma,
	colon,
	lparen,
	rparen,
	lbrace,
	rbrace,
	lbracket,
	rbracket,
	
	// reserved words
	fn,
	let,
	_true,
	_false,
	_if,
	_else,
	_return,

	// #
	comment,
};

struct Token {
	TokenType type;

	std::string str;
	i64 inum;
	float fnum;
};

std::vector<Token> tokenize(std::string line);

} // namespace shiba
