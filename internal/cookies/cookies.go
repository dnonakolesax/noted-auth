package cookies

import (
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

func SetupAccessCookies(ctx *fasthttp.RequestCtx, tokenDTO model.TokenDTO) {
	atCookie := fasthttp.Cookie{}
	atCookie.SetKey(consts.ATCookieKey)
	atCookie.SetValue(tokenDTO.AccessToken)
	atCookie.SetMaxAge(tokenDTO.ExpiresIn)
	atCookie.SetHTTPOnly(true)
	atCookie.SetSecure(true)
	atCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	atCookie.SetPath("/")

	rtCookie := fasthttp.Cookie{}
	rtCookie.SetKey(consts.RTCookieKey)
	rtCookie.SetValue(tokenDTO.RefreshToken)
	rtCookie.SetMaxAge(tokenDTO.RefreshExp)
	rtCookie.SetHTTPOnly(true)
	rtCookie.SetSecure(true)
	rtCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	rtCookie.SetPath("/")

	idtCookie := fasthttp.Cookie{}
	idtCookie.SetKey(consts.IDTCookieKey)
	idtCookie.SetValue(tokenDTO.IDToken)
	idtCookie.SetHTTPOnly(true)
	idtCookie.SetSecure(true)
	idtCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	idtCookie.SetPath("/")

	ctx.Response.Header.SetCookie(&atCookie)
	ctx.Response.Header.SetCookie(&rtCookie)
	ctx.Response.Header.SetCookie(&idtCookie)
	ctx.Request.Header.SetCookie(consts.ATCookieKey, tokenDTO.AccessToken)
	ctx.Request.Header.SetCookie(consts.RTCookieKey, tokenDTO.RefreshToken)
	ctx.Request.Header.SetCookie(consts.IDTCookieKey, tokenDTO.IDToken)
}

func EraseAccessCookies(ctx *fasthttp.RequestCtx) {
	atCookie := fasthttp.Cookie{}
	atCookie.SetKey(consts.ATCookieKey)
	atCookie.SetValue("")
	atCookie.SetMaxAge(-1)
	atCookie.SetHTTPOnly(true)
	atCookie.SetSecure(true)
	atCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	atCookie.SetPath("/")

	rtCookie := fasthttp.Cookie{}
	rtCookie.SetKey(consts.RTCookieKey)
	rtCookie.SetValue("")
	rtCookie.SetMaxAge(-1)
	rtCookie.SetHTTPOnly(true)
	rtCookie.SetSecure(true)
	rtCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	rtCookie.SetPath("/")

	idtCookie := fasthttp.Cookie{}
	idtCookie.SetKey(consts.IDTCookieKey)
	idtCookie.SetValue("")
	rtCookie.SetMaxAge(-1)
	idtCookie.SetHTTPOnly(true)
	idtCookie.SetSecure(true)
	idtCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	idtCookie.SetPath("/")

	ctx.Response.Header.SetCookie(&atCookie)
	ctx.Response.Header.SetCookie(&rtCookie)
	ctx.Response.Header.SetCookie(&idtCookie)
}
