---@class GoodsModule
local GoodsModule = class('GoodsModule')


---@return GoodsModule
function GoodsModule.new()
end


function GoodsModule:init()
    self.goodsId   = 0
    self.goodsType = 0
    self.goodsNum  = 0
    self.goodsName = ''
end


function GoodsModule:isNew()
end


return GoodsModule
