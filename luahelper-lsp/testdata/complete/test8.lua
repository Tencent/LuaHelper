---@class lose_label_data @用户回流标签数据
local lose_label_data = {
    open_flag = 1, -- 开启的标记
    open_time = now, -- 回流标记初始化的时间，后面在7天内，累计登录的前3日，每次都给自己活跃好友发送信息
    continue_day_num = 1, -- 会回流后，7天内累计登录的天数
    ---@type number[] @邀请自己组队的列表，最多存储10个
    invite_self_list = { 
        -- [1] = uid1,  -- 存储好友的uid1
        -- [2] = uid2,  -- 存储好友的uid2
    },
    get_flag = 0 -- 拉取状态值，整个系统依赖于好友是否在线和好友的亲密度；存储的是 get_interface_array 位数据
}

---@type lose_label_data
local lose_label_data = {}



