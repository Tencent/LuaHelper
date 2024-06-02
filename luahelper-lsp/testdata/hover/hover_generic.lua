
---@class UnitObject
local UnitClass = class('UnitObject')


---@class HeroObject : UnitObject
local HeroObject = class('HeroObject', UnitObject)


---@class EmptyObject : UnitObject
local EmptyObject = class('EmptyObject', UnitObject)


---@generic T
---@param classType T
---@return T
function GoodsBuild(classType)
end


---@generic T
---@param className `T`
---@return T
function GoodsBuildByName(className)
end


---@generic T
---@param classType T
---@return T[]
function GoodsBuilds(classType)
end


---@generic T
---@param className `T`
---@return T[]
function GoodsBuildsByName(className)
end


local nameLocal   = 'GoodsModule'
local nameGlobal  = GoodsDefine.BaseGoods
local classLocal  = require('hover_generic_class')
local classGlobal = GoodsClass.BaseClass
-- 全局方法定义测试
local cloneGoods1 = clone(EmptyObject)
local cloneGoods2 = clone(classLocal)
local cloneGoods3 = clone(classGlobal)
local cloneGoods4 = clone(GoodsClass.BaseClass)
-- 局部方法定义测试
local classGoods1 = GoodsBuild(classLocal)
local classGoods2 = GoodsBuild(classGlobal)
-- 字符串泛型定义测试
local nameGoods1 = GoodsBuildByName('GoodsModule')
local nameGoods2 = GoodsBuildByName(nameLocal)
local nameGoods3 = GoodsBuildByName(nameGlobal)
local nameGoods4 = GoodsBuildByName(GoodsDefine.BaseGoods)
local nameGoods5 = createByName('GoodsModule')
local nameGoods6 = createByName(nameLocal)
local nameGoods7 = createByName(nameGlobal)
local nameGoods8 = createByName(GoodsDefine.BaseGoods)
-- 数组泛型定义测试
local goodsList1 = GoodsBuilds(classLocal)
local goodsList2 = GoodsBuilds(GoodsClass.BaseClass)
local goodsList3 = GoodsBuildsByName('GoodsModule')
local goodsList4 = GoodsBuildsByName(nameGlobal)
local goodsList5 = GoodsBuildsByName(GoodsDefine.BaseGoods)
local goodsList6 = createListByName('GoodsModule')
local goodsList7 = createListByName(nameGlobal)
local goodsList8 = createListByName(GoodsDefine.BaseGoods)
