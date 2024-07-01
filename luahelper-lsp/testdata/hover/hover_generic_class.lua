---@class GoodsModule
local GoodsModule = class('GoodsModule')


---@return GoodsModule
function GoodsModule.new()
end


---@return GoodsModule
function GoodsModule.GetInstance()
end


function GoodsModule:init()
    self.goodsId   = 0
    self.goodsType = 0
    self.goodsNum  = 0
    self.goodsName = ''
end


---@return boolean
function GoodsModule:isNew()
end


---@generic T
---@param goodsObj T
---@return T
function GoodsModule:CloneGoods(goodsObj)
end


---@generic T
---@param goodsObj T
---@return T[]
function GoodsModule:CloneGoodsList(goodsObj)
end


---@generic T
---@param goodsName `T`
---@return T
function GoodsModule:CreateGoods(goodsName)
end


---@generic T
---@param goodsName `T`
---@return T[]
function GoodsModule:CreateGoodsList(goodsName)
end


return GoodsModule
