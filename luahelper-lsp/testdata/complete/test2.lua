---@class uiButton
local uiButton = class('uiButton')

---@return uiButton
function uiButton:new()
    return self
end

---@return uiButton
function uiButton:setX(x)
    self.x_ = x
    return self
end

---@return uiButton
function uiButton:setY(y)
    self.y_ = y
    self.setY():setY():setY():setY()
    return self
end

local btn1 = uiButton.new()                    -- 返回类型 uiButton

local btn2 = uiButton.new():setX(10)           -- 返回类型 any

local btn3 = uiButton.new():setX(10):setY(10)  -- 返回类型 any

_G.aaa = uiButton.new()
local btn4 = _G.aaa.setX().setX():setX():setX():setX()

local btn5 = uiButton:setX(1)

---@return uiButton
function setYa(y)
    self.y_ = y
    uiButton.setY():setY():setY():setY().setX()
    return self
end

local btn6 = setYa();
local btn7 = setYa().setX();


---@return uiButton
function uiButton:setY1(y)
    local btn8 = self:setX()
    local btn9 = self:setY()
       
end
           
           