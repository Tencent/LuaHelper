local now 

---@class lose_battle_backflow 
---@field open_flag number 
---@field open_time number 
---@field last_update_time number 
---@field continue_day_num number 
local back_flow = {
    open_flag = 1, 
    open_time = now, 
    last_update_time = now, 
    continue_day_num = 1
}

print(back_flow)