package cookies

import (
	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/model"
	"github.com/valyala/fasthttp"
)

func SetupAccessCookies(ctx *fasthttp.RequestCtx, tokenDTO model.TokenDTO) {
	atCookie := fasthttp.Cookie{}
	atCookie.SetKey(consts.ATCookieKey)
	atCookie.SetValue(tokenDTO.AccessToken)
	atCookie.SetMaxAge(tokenDTO.ExpiresIn)
	atCookie.SetHTTPOnly(true)
	atCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	atCookie.SetPath("/")

	rtCookie := fasthttp.Cookie{}
	rtCookie.SetKey(consts.RTCookieKey)
	rtCookie.SetValue(tokenDTO.RefreshToken)
	rtCookie.SetMaxAge(tokenDTO.RefreshExp)
	rtCookie.SetHTTPOnly(true)
	rtCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	rtCookie.SetPath("/")

	ctx.Response.Header.SetCookie(&atCookie)
	ctx.Response.Header.SetCookie(&rtCookie)
}
