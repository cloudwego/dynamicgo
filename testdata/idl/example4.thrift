include "base.thrift"
namespace go example4

struct Favorite {
    1: required i32 Uid,
    2: required list<string> Stuffs,
}

struct GetFavoriteReq {
    1: required i32 Id (api.query = "id"),
    255: required Favorite Base,
}
struct GetFavoriteResp {
    1: required Favorite Favorite (api.key="favorite"),
    255: required base.BaseResp BaseResp,
}

service GetFavoriteService {
    GetFavoriteResp GetFavoriteMethod(1: GetFavoriteReq req) (api.get="/v1/GetFavorite/:id"),
}
