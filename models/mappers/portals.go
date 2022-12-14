package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/portals"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func MapPortalPositionRequest(dimension entities.Dimension, server entities.Server, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_PORTAL_POSITION_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		PortalPositionRequest: &amqp.PortalPositionRequest{
			DimensionId: dimension.Id,
			ServerId:    server.Id,
		},
	}
}

func MapToEmbed(portal *amqp.PortalPositionAnswer_PortalPosition, portalService portals.PortalService, locale amqp.RabbitMQMessage_Language) *discordgo.MessageEmbed {
	lg := constants.MapAmqpLocale(locale)
	dimension, found := portalService.GetDimension(portal.DimensionId)
	if !found {
		log.Warn().Str(constants.LogEntity, portal.DimensionId).Msgf("Cannot find dimension based on ID sent internally, continuing with empty dimension")
		dimension = entities.Dimension{Id: portal.DimensionId}
	}

	embed := discordgo.MessageEmbed{
		Title:     translators.GetEntityLabel(dimension, lg),
		Color:     dimension.Color,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: dimension.Icon},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    i18n.Get(lg, "pos.embed.footer", i18n.Vars{"source": portal.Source.Name}),
			IconURL: portal.Source.Icon,
		},
	}

	if portal.Position != nil {
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:   i18n.Get(lg, "pos.embed.position.name"),
				Value:  i18n.Get(lg, "pos.embed.position.value", i18n.Vars{"position": portal.Position}),
				Inline: true,
			},
			{
				Name:   i18n.Get(lg, "pos.embed.uses.name"),
				Value:  i18n.Get(lg, "pos.embed.uses.value", i18n.Vars{"uses": portal.RemainingUses}),
				Inline: true,
			},
			{
				Name: i18n.Get(lg, "pos.embed.hunters.name"),
				Value: i18n.Get(lg, "pos.embed.hunters.value", i18n.Vars{
					"createdBy": portal.CreatedBy, "createdAt": portal.CreatedAt,
					"updatedBy": portal.UpdatedBy, "updatedAt": portal.UpdatedAt,
				}),
				Inline: true,
			},
		}

		if portal.Position.ConditionalTransport != nil {
			embed.Fields = append(embed.Fields, mapTransportToEmbed(portal.Position.ConditionalTransport, portalService, lg, true))
		}

		embed.Fields = append(embed.Fields, mapTransportToEmbed(portal.Position.Transport, portalService, lg, false))

	} else {
		embed.Description = i18n.Get(lg, "pos.embed.unknown")
	}

	return &embed
}

func mapTransportToEmbed(transport *amqp.PortalPositionAnswer_PortalPosition_Position_Transport,
	portalService portals.PortalService, lg discordgo.Locale, isConditional bool) *discordgo.MessageEmbedField {

	transportType, found := portalService.GetTransportType(transport.TypeId)
	if !found {
		log.Warn().
			Str(constants.LogEntity, transport.TypeId).
			Msgf("Cannot find transport type based on ID sent internally, continuing with empty transport")
		transportType = entities.TransportType{Id: transport.TypeId}
	}

	area, found := portalService.GetArea(transport.AreaId)
	if !found {
		log.Warn().
			Str(constants.LogEntity, transport.AreaId).
			Msgf("Cannot find area based on ID sent internally, continuing with empty area")
		area = entities.Area{Id: transport.AreaId}
	}

	subArea, found := portalService.GetSubArea(transport.SubAreaId)
	if !found {
		log.Warn().
			Str(constants.LogEntity, transport.SubAreaId).
			Msgf("Cannot find sub area based on ID sent internally, continuing with empty sub area")
		subArea = entities.SubArea{Id: transport.SubAreaId}
	}

	return &discordgo.MessageEmbedField{
		Name: i18n.Get(lg, "pos.embed.transport.name", i18n.Vars{
			"type":        translators.GetEntityLabel(transportType, lg),
			"conditional": isConditional,
		}),
		Value: i18n.Get(lg, "pos.embed.transport.value", i18n.Vars{
			"area":    translators.GetEntityLabel(area, lg),
			"subArea": translators.GetEntityLabel(subArea, lg),
			"x":       transport.X,
			"y":       transport.Y,
		}),
		Inline: false,
	}
}
