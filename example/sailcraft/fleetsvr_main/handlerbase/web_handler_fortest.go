package handlerbase

type WebHandlerTest struct {
	WebHandler
}

func (this *WebHandlerTest) Test() (int, error) {
	this.Response.ResData.Params = this.Request.ReqData.Params
	return 0, nil
}

func (this *WebHandlerTest) Test1() (int, error) {
	this.Response.ResData.Params = "This is test1"
	return 0, nil
}

func (this *WebHandlerTest) Test2() (int, error) {
	this.Response.ResData.Params = "========================"
	return 0, nil
}
